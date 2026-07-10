package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"dentaldesk/apps/api/internal/auth"
	"dentaldesk/apps/api/internal/config"
	"dentaldesk/apps/api/internal/domain"
	"dentaldesk/apps/api/internal/notifications"
	"dentaldesk/apps/api/internal/platform/httpx"
	"dentaldesk/apps/api/internal/practices"
	"dentaldesk/apps/api/internal/ratelimit"
	"dentaldesk/apps/api/internal/store"
	"dentaldesk/apps/api/internal/voice"
)

type Server struct {
	cfg           config.Config
	logger        *slog.Logger
	store         store.Store
	auth          *auth.Service
	practices     *practices.Service
	voice         *voice.Service
	notifications *notifications.Service
	authLimiter   *ratelimit.Limiter
	mux           *http.ServeMux
}

func NewServer(cfg config.Config, logger *slog.Logger) (*Server, error) {
	st, err := newStore(cfg)
	if err != nil {
		return nil, err
	}
	s := &Server{
		cfg:           cfg,
		logger:        logger,
		store:         st,
		auth:          auth.NewService(st, cfg.SessionSecret, cfg.AppEnv == "production"),
		practices:     practices.NewService(st),
		voice:         voice.NewService(st, cfg.VAPIWebhookSecret),
		notifications: notifications.NewService(logger, cfg),
		authLimiter:   ratelimit.New(10, 10*time.Minute),
		mux:           http.NewServeMux(),
	}
	s.routes()
	return s, nil
}

func newStore(cfg config.Config) (store.Store, error) {
	switch cfg.StoreDriver {
	case "", "memory":
		return store.NewMemoryStore(), nil
	case "postgres":
		if cfg.DatabaseURL == "" {
			return nil, fmt.Errorf("DATABASE_URL is required when STORE_DRIVER=postgres")
		}
		return store.NewPostgresStore(context.Background(), cfg.DatabaseURL)
	default:
		return nil, fmt.Errorf("unsupported store driver %q", cfg.StoreDriver)
	}
}

func (s *Server) Handler() http.Handler {
	return s.withCORS(s.withCSRF(s.mux))
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.health)
	s.mux.HandleFunc("GET /v1/csrf", s.csrf)
	s.mux.HandleFunc("POST /v1/auth/register", s.register)
	s.mux.HandleFunc("POST /v1/auth/login", s.login)
	s.mux.HandleFunc("POST /v1/auth/logout", s.logout)
	s.mux.HandleFunc("POST /v1/auth/verify-email", s.verifyEmail)
	s.mux.HandleFunc("POST /v1/auth/request-password-reset", s.requestPasswordReset)
	s.mux.HandleFunc("POST /v1/auth/reset-password", s.resetPassword)
	s.mux.HandleFunc("GET /v1/me", s.requireUser(s.me))
	s.mux.HandleFunc("GET /v1/practices", s.requireUser(s.listPractices))
	s.mux.HandleFunc("POST /v1/practices", s.requireUser(s.createPractice))
	s.mux.HandleFunc("GET /v1/practices/{practiceID}/assistant-config", s.requirePractice(s.getAssistantConfig))
	s.mux.HandleFunc("PATCH /v1/practices/{practiceID}/assistant-config", s.requirePermission("assistant_config:update", s.updateAssistantConfig))
	s.mux.HandleFunc("GET /v1/practices/{practiceID}/voice-provider", s.requirePractice(s.getVoiceProviderConfig))
	s.mux.HandleFunc("PATCH /v1/practices/{practiceID}/voice-provider", s.requirePermission("assistant_config:update", s.updateVoiceProviderConfig))
	s.mux.HandleFunc("GET /v1/practices/{practiceID}/calendar-config", s.requirePractice(s.getCalendarConfig))
	s.mux.HandleFunc("PATCH /v1/practices/{practiceID}/calendar-config", s.requirePermission("calendar:update", s.updateCalendarConfig))
	s.mux.HandleFunc("POST /v1/practices/{practiceID}/calendar/oauth/start", s.requirePermission("calendar:update", s.startGoogleCalendarOAuth))
	s.mux.HandleFunc("GET /v1/calendar/oauth/callback", s.requireUser(s.googleCalendarOAuthCallback))
	s.mux.HandleFunc("POST /v1/practices/{practiceID}/calendar/events", s.requirePermission("calendar:update", s.createGoogleCalendarEvent))
	s.mux.HandleFunc("GET /v1/practices/{practiceID}/billing", s.requirePermission("billing:read", s.getBillingSubscription))
	s.mux.HandleFunc("PATCH /v1/practices/{practiceID}/billing", s.requirePermission("billing:update", s.updateBillingSubscription))
	s.mux.HandleFunc("POST /v1/practices/{practiceID}/billing/checkout-session", s.requirePermission("billing:update", s.createStripeCheckoutSession))
	s.mux.HandleFunc("POST /v1/billing/stripe/webhook", s.stripeWebhook)
	s.mux.HandleFunc("POST /v1/practices/{practiceID}/voice-test-call", s.requirePermission("assistant_config:update", s.runVoiceTestCall))
	s.mux.HandleFunc("GET /v1/practices/{practiceID}/call-summaries", s.requirePractice(s.listCallSummaries))
	s.mux.HandleFunc("GET /v1/practices/{practiceID}/appointment-requests", s.requirePractice(s.listAppointmentRequests))
	s.mux.HandleFunc("PATCH /v1/practices/{practiceID}/appointment-requests/{requestID}", s.requirePermission("appointment:update", s.updateAppointmentRequest))
	s.mux.HandleFunc("GET /v1/practices/{practiceID}/locations", s.requirePermission("location:update", s.listLocations))
	s.mux.HandleFunc("POST /v1/practices/{practiceID}/locations", s.requirePermission("location:create", s.createLocation))
	s.mux.HandleFunc("GET /v1/practices/{practiceID}/roles", s.requirePermission("role:update", s.listRoles))
	s.mux.HandleFunc("POST /v1/practices/{practiceID}/roles", s.requirePermission("role:create", s.createRole))
	s.mux.HandleFunc("GET /v1/practices/{practiceID}/members", s.requirePermission("member:update", s.listMembers))
	s.mux.HandleFunc("GET /v1/practices/{practiceID}/activity", s.requirePractice(s.listActivity))
	s.mux.HandleFunc("POST /v1/practices/{practiceID}/invites", s.requirePermission("member:invite", s.createInvite))
	s.mux.HandleFunc("GET /v1/practices/{practiceID}/invites", s.requirePermission("member:invite", s.listInvites))
	s.mux.HandleFunc("PATCH /v1/practices/{practiceID}/members/{userID}", s.requirePermission("member:update", s.updateMember))
	s.mux.HandleFunc("DELETE /v1/practices/{practiceID}/members/{userID}", s.requirePermission("member:disable", s.disableMember))
	s.mux.HandleFunc("GET /v1/invites/{token}", s.getInvite)
	s.mux.HandleFunc("POST /v1/invites/{token}/accept", s.acceptInvite)
	s.mux.HandleFunc("POST /v1/voice/bootstrap", s.voiceBootstrap)
	s.mux.HandleFunc("POST /v1/voice/practice-info", s.voicePracticeInfo)
	s.mux.HandleFunc("POST /v1/voice/appointment-request", s.voiceAppointmentRequest)
	s.mux.HandleFunc("POST /v1/voice/call-summary", s.voiceCallSummary)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) csrf(w http.ResponseWriter, r *http.Request) {
	token, err := s.auth.NewCSRFToken()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not create csrf token")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CSRFCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.cfg.AppEnv == "production",
		Expires:  time.Now().UTC().Add(12 * time.Hour),
	})
	httpx.JSON(w, http.StatusOK, map[string]string{"csrfToken": token})
}

type authRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	if !s.allowAuthAttempt(r, "register") {
		httpx.Error(w, http.StatusTooManyRequests, "too many attempts")
		return
	}
	var req authRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request")
		return
	}
	user, token, err := s.auth.Register(r.Context(), req.Email, req.DisplayName, req.Password)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	s.auth.SetSessionCookie(w, token)
	s.sendVerificationEmail(r, user)
	httpx.JSON(w, http.StatusCreated, user)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if !s.allowAuthAttempt(r, "login") {
		httpx.Error(w, http.StatusTooManyRequests, "too many attempts")
		return
	}
	var req authRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request")
		return
	}
	user, token, err := s.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	s.auth.SetSessionCookie(w, token)
	httpx.JSON(w, http.StatusOK, user)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	s.auth.ClearSessionCookie(w)
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) verifyEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := httpx.Decode(r, &req); err != nil || strings.TrimSpace(req.Token) == "" {
		httpx.Error(w, http.StatusBadRequest, "token is required")
		return
	}
	user, ok := s.auth.VerifyEmail(r.Context(), req.Token)
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "invalid or expired token")
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

func (s *Server) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if !s.allowAuthAttempt(r, "password-reset") {
		httpx.Error(w, http.StatusTooManyRequests, "too many attempts")
		return
	}
	var req struct {
		Email string `json:"email"`
	}
	if err := httpx.Decode(r, &req); err != nil || strings.TrimSpace(req.Email) == "" {
		httpx.Error(w, http.StatusBadRequest, "email is required")
		return
	}
	user, token, ok, err := s.auth.CreatePasswordReset(r.Context(), req.Email)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not create reset token")
		return
	}
	if ok {
		resetURL := s.cfg.WebBaseURL + "/reset-password?token=" + token
		if err := s.notifications.SendPasswordReset(r.Context(), user.Email, resetURL); err != nil {
			s.logger.Error("password reset email failed", "user_id", user.ID, "error", err)
		}
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) resetPassword(w http.ResponseWriter, r *http.Request) {
	if !s.allowAuthAttempt(r, "reset-password") {
		httpx.Error(w, http.StatusTooManyRequests, "too many attempts")
		return
	}
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := httpx.Decode(r, &req); err != nil || strings.TrimSpace(req.Token) == "" {
		httpx.Error(w, http.StatusBadRequest, "token is required")
		return
	}
	user, ok, err := s.auth.ResetPassword(r.Context(), req.Token, req.Password)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "invalid or expired token")
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

func (s *Server) me(w http.ResponseWriter, r *http.Request, userID string) {
	httpx.JSON(w, http.StatusOK, map[string]string{"userId": userID})
}

func (s *Server) listPractices(w http.ResponseWriter, r *http.Request, userID string) {
	httpx.JSON(w, http.StatusOK, s.practices.ListForUser(r.Context(), userID))
}

func (s *Server) createPractice(w http.ResponseWriter, r *http.Request, userID string) {
	var req struct {
		Name string `json:"name"`
	}
	if err := httpx.Decode(r, &req); err != nil || strings.TrimSpace(req.Name) == "" {
		httpx.Error(w, http.StatusBadRequest, "practice name is required")
		return
	}
	practice := s.practices.Create(r.Context(), userID, req.Name)
	s.audit(practice.ID, "user", userID, "practice.created", "practice", practice.ID, map[string]string{"name": practice.Name})
	httpx.JSON(w, http.StatusCreated, practice)
}

func (s *Server) getAssistantConfig(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	config, ok := s.store.GetAssistantConfig(practiceID)
	if !ok {
		httpx.Error(w, http.StatusNotFound, "assistant config not found")
		return
	}
	httpx.JSON(w, http.StatusOK, config)
}

func (s *Server) updateAssistantConfig(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	current, ok := s.store.GetAssistantConfig(practiceID)
	if !ok {
		httpx.Error(w, http.StatusNotFound, "assistant config not found")
		return
	}
	var req domain.AssistantConfig
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid assistant config")
		return
	}
	if strings.TrimSpace(req.Greeting) != "" {
		current.Greeting = strings.TrimSpace(req.Greeting)
	}
	current.EscalationPhone = strings.TrimSpace(req.EscalationPhone)
	current.NotificationEmail = strings.TrimSpace(req.NotificationEmail)
	if req.Settings != nil {
		current.Settings = req.Settings
	}
	current.PracticeID = practiceID
	updated := s.store.SaveAssistantConfig(current)
	s.audit(practiceID, "user", userID, "assistant_config.updated", "assistant_config", updated.ID, map[string]string{"notificationEmail": updated.NotificationEmail})
	httpx.JSON(w, http.StatusOK, updated)
}

func (s *Server) getVoiceProviderConfig(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	config, ok := s.store.GetVoiceProviderConfig(practiceID)
	if !ok {
		config = s.store.SaveVoiceProviderConfig(domain.VoiceProviderConfig{
			PracticeID:    practiceID,
			Provider:      "vapi",
			WebhookStatus: "not_configured",
		})
	}
	httpx.JSON(w, http.StatusOK, config)
}

func (s *Server) updateVoiceProviderConfig(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	current, ok := s.store.GetVoiceProviderConfig(practiceID)
	if !ok {
		current = domain.VoiceProviderConfig{
			PracticeID:    practiceID,
			Provider:      "vapi",
			WebhookStatus: "not_configured",
		}
	}
	var req domain.VoiceProviderConfig
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid voice provider config")
		return
	}
	current.PracticeID = practiceID
	current.Provider = normalizeProvider(req.Provider)
	current.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
	current.AssistantID = strings.TrimSpace(req.AssistantID)
	current.WebhookStatus = webhookStatus(current)
	updated := s.store.SaveVoiceProviderConfig(current)
	s.audit(practiceID, "user", userID, "voice_provider.updated", "voice_provider", updated.ID, map[string]string{"provider": updated.Provider, "webhookStatus": updated.WebhookStatus})
	httpx.JSON(w, http.StatusOK, updated)
}

func (s *Server) getCalendarConfig(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	config, ok := s.store.GetCalendarConfig(practiceID)
	if !ok {
		config = s.store.SaveCalendarConfig(domain.CalendarConfig{
			PracticeID: practiceID,
			Mode:       "request_only",
			Provider:   "none",
			Timezone:   "America/New_York",
			Status:     "not_configured",
		})
	}
	httpx.JSON(w, http.StatusOK, config)
}

func (s *Server) updateCalendarConfig(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	current, ok := s.store.GetCalendarConfig(practiceID)
	if !ok {
		current = domain.CalendarConfig{PracticeID: practiceID}
	}
	var req domain.CalendarConfig
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid calendar config")
		return
	}
	current.Mode = normalizeCalendarMode(req.Mode)
	current.Provider = normalizeCalendarProvider(req.Provider, current.Mode)
	current.BookingURL = strings.TrimSpace(req.BookingURL)
	current.CalendarID = strings.TrimSpace(req.CalendarID)
	current.Timezone = strings.TrimSpace(req.Timezone)
	current.Instructions = strings.TrimSpace(req.Instructions)
	current.Status = calendarStatus(current)
	updated := s.store.SaveCalendarConfig(current)
	s.audit(practiceID, "user", userID, "calendar_config.updated", "calendar_config", updated.ID, map[string]string{"mode": updated.Mode, "status": updated.Status})
	httpx.JSON(w, http.StatusOK, updated)
}

func (s *Server) getBillingSubscription(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	subscription, ok := s.store.GetBillingSubscription(practiceID)
	if !ok {
		subscription = s.store.SaveBillingSubscription(domain.BillingSubscription{
			PracticeID: practiceID,
		})
	}
	httpx.JSON(w, http.StatusOK, subscription)
}

func (s *Server) updateBillingSubscription(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	current, ok := s.store.GetBillingSubscription(practiceID)
	if !ok {
		current = domain.BillingSubscription{PracticeID: practiceID}
	}
	var req domain.BillingSubscription
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid billing subscription")
		return
	}
	current.Plan = normalizeBillingPlan(req.Plan)
	current.Status = normalizeBillingStatus(req.Status)
	current.IncludedMinutes = req.IncludedMinutes
	current.OverageCents = req.OverageCents
	current.StripeCustomerID = strings.TrimSpace(req.StripeCustomerID)
	current.StripeSubscriptionID = strings.TrimSpace(req.StripeSubscriptionID)
	current.TrialEndsAt = req.TrialEndsAt
	current.CurrentPeriodEndsAt = req.CurrentPeriodEndsAt
	updated := s.store.SaveBillingSubscription(current)
	s.audit(practiceID, "user", userID, "billing.updated", "billing_subscription", updated.ID, map[string]string{"plan": updated.Plan, "status": updated.Status})
	httpx.JSON(w, http.StatusOK, updated)
}

func (s *Server) runVoiceTestCall(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	var req struct {
		CallerName    string `json:"callerName"`
		CallerPhone   string `json:"callerPhone"`
		RequestType   string `json:"requestType"`
		PreferredTime string `json:"preferredTime"`
		Insurance     string `json:"insurance"`
		Notes         string `json:"notes"`
		Urgency       string `json:"urgency"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid test call")
		return
	}
	if strings.TrimSpace(req.CallerName) == "" {
		req.CallerName = "Test Caller"
	}
	if strings.TrimSpace(req.CallerPhone) == "" {
		req.CallerPhone = "+1555010100"
	}
	if strings.TrimSpace(req.RequestType) == "" {
		req.RequestType = "new patient cleaning"
	}
	if strings.TrimSpace(req.PreferredTime) == "" {
		req.PreferredTime = "weekday morning"
	}
	if strings.TrimSpace(req.Urgency) == "" {
		req.Urgency = "routine"
	}

	call := s.store.CreateCall(domain.CallSession{
		PracticeID:     practiceID,
		Provider:       "test_harness",
		ProviderCallID: "test-" + time.Now().UTC().Format("20060102150405"),
		CallerPhone:    req.CallerPhone,
		Status:         "completed",
		StartedAt:      time.Now().UTC(),
	})
	appointment := s.voice.CreateAppointmentRequest(r.Context(), domain.AppointmentRequest{
		PracticeID:    practiceID,
		CallSessionID: call.ID,
		CallerName:    req.CallerName,
		CallerPhone:   req.CallerPhone,
		RequestType:   req.RequestType,
		PreferredTime: req.PreferredTime,
		Insurance:     req.Insurance,
		Notes:         req.Notes,
		Status:        "new",
	})
	summary := s.voice.CreateCallSummary(r.Context(), domain.CallSummary{
		CallSessionID:  call.ID,
		PracticeID:     practiceID,
		CallerName:     req.CallerName,
		Reason:         req.RequestType,
		Urgency:        req.Urgency,
		AIAction:       "Created appointment request from local voice test harness",
		FollowUpNeeded: "Staff should confirm the requested appointment time.",
		Summary:        req.CallerName + " requested " + req.RequestType + " and prefers " + req.PreferredTime + ".",
	})
	if config, ok := s.store.GetAssistantConfig(practiceID); ok {
		if err := s.notifications.SendAppointmentRequestEmail(r.Context(), config.NotificationEmail, appointment); err != nil {
			s.logger.Error("test appointment notification failed", "practice_id", practiceID, "error", err)
		}
		if err := s.notifications.SendCallSummaryEmail(r.Context(), config.NotificationEmail, summary); err != nil {
			s.logger.Error("test summary notification failed", "practice_id", practiceID, "error", err)
		}
	}
	s.audit(practiceID, "user", userID, "voice_test_call.created", "call_session", call.ID, map[string]string{"callerName": req.CallerName, "requestType": req.RequestType})
	httpx.JSON(w, http.StatusCreated, map[string]any{
		"call":    appointment,
		"summary": summary,
	})
}

func (s *Server) listCallSummaries(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	httpx.JSON(w, http.StatusOK, s.store.ListSummaries(practiceID))
}

func (s *Server) listAppointmentRequests(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	httpx.JSON(w, http.StatusOK, s.store.ListAppointments(practiceID))
}

func (s *Server) updateAppointmentRequest(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	var req struct {
		Status    string `json:"status"`
		StaffNote string `json:"staffNote"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid appointment update")
		return
	}
	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	if req.Status != "" && !validAppointmentStatus(req.Status) {
		httpx.Error(w, http.StatusBadRequest, "invalid appointment status")
		return
	}
	requestID := r.PathValue("requestID")
	appointment, ok := s.store.UpdateAppointmentRequest(practiceID, requestID, req.Status, strings.TrimSpace(req.StaffNote))
	if !ok {
		httpx.Error(w, http.StatusNotFound, "appointment request not found")
		return
	}
	s.audit(practiceID, "user", userID, "appointment_request.updated", "appointment_request", appointment.ID, map[string]string{"status": appointment.Status})
	httpx.JSON(w, http.StatusOK, appointment)
}

func (s *Server) listLocations(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	httpx.JSON(w, http.StatusOK, s.store.ListLocations(practiceID))
}

func (s *Server) createLocation(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	var req domain.Location
	if err := httpx.Decode(r, &req); err != nil || strings.TrimSpace(req.Name) == "" {
		httpx.Error(w, http.StatusBadRequest, "location name is required")
		return
	}
	req.PracticeID = practiceID
	location := s.store.CreateLocation(req)
	s.audit(practiceID, "user", userID, "location.created", "location", location.ID, map[string]string{"name": location.Name})
	httpx.JSON(w, http.StatusCreated, location)
}

func (s *Server) listRoles(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	httpx.JSON(w, http.StatusOK, s.store.ListRoles(practiceID))
}

func (s *Server) createRole(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}
	if err := httpx.Decode(r, &req); err != nil || strings.TrimSpace(req.Name) == "" {
		httpx.Error(w, http.StatusBadRequest, "role name is required")
		return
	}
	if _, exists := s.store.GetRole(practiceID, req.Name); exists {
		httpx.Error(w, http.StatusConflict, "role already exists")
		return
	}
	role := s.store.CreateRole(practiceID, req.Name, req.Description, req.Permissions)
	s.audit(practiceID, "user", userID, "role.created", "role", role.ID, map[string]string{"name": role.Name})
	httpx.JSON(w, http.StatusCreated, role)
}

func (s *Server) listMembers(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	httpx.JSON(w, http.StatusOK, s.store.ListMembers(practiceID))
}

func (s *Server) listActivity(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	httpx.JSON(w, http.StatusOK, s.store.ListAuditLogs(practiceID, 50))
}

func (s *Server) createInvite(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := httpx.Decode(r, &req); err != nil || strings.TrimSpace(req.Email) == "" {
		httpx.Error(w, http.StatusBadRequest, "invite email is required")
		return
	}
	if req.Role == "" {
		req.Role = "staff"
	}
	if _, ok := s.store.GetRole(practiceID, req.Role); !ok {
		httpx.Error(w, http.StatusBadRequest, "unknown role")
		return
	}
	token, err := s.auth.NewPublicToken()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not create invite")
		return
	}
	invite := s.store.CreateInvite(practiceID, req.Email, req.Role, userID, s.auth.HashPublicToken(token), time.Now().UTC().Add(7*24*time.Hour))
	inviteURL := s.cfg.WebBaseURL + "/accept-invite?token=" + token
	emailSent := true
	emailError := ""
	if err := s.notifications.SendInviteEmail(r.Context(), req.Email, req.Role, inviteURL); err != nil {
		emailSent = false
		emailError = "invite created but email could not be sent"
		s.logger.Error("invite email failed", "practice_id", practiceID, "invite_id", invite.ID, "error", err)
	}
	s.audit(practiceID, "user", userID, "invite.created", "practice_invite", invite.ID, map[string]string{"email": invite.Email, "role": invite.Role, "emailSent": fmt.Sprint(emailSent)})
	type response struct {
		domain.PracticeInvite
		InviteURL  string `json:"inviteUrl"`
		EmailSent  bool   `json:"emailSent"`
		EmailError string `json:"emailError,omitempty"`
	}
	httpx.JSON(w, http.StatusCreated, response{
		PracticeInvite: invite,
		InviteURL:      inviteURL,
		EmailSent:      emailSent,
		EmailError:     emailError,
	})
}

func (s *Server) listInvites(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	httpx.JSON(w, http.StatusOK, s.store.ListInvites(practiceID))
}

func (s *Server) getInvite(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	invite, ok := s.store.GetInviteByTokenHash(s.auth.HashPublicToken(token))
	if !ok {
		httpx.Error(w, http.StatusNotFound, "invite not found or expired")
		return
	}
	httpx.JSON(w, http.StatusOK, invite)
}

func (s *Server) acceptInvite(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	tokenHash := s.auth.HashPublicToken(token)
	invite, ok := s.store.GetInviteByTokenHash(tokenHash)
	if !ok {
		httpx.Error(w, http.StatusNotFound, "invite not found or expired")
		return
	}
	var req struct {
		DisplayName string `json:"displayName"`
		Password    string `json:"password"`
	}
	if err := httpx.Decode(r, &req); err != nil || strings.TrimSpace(req.Password) == "" {
		httpx.Error(w, http.StatusBadRequest, "password is required")
		return
	}
	user, err := s.auth.CreateUser(r.Context(), invite.Email, req.DisplayName, req.Password)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "could not create invited user")
		return
	}
	accepted, ok := s.store.AcceptInvite(tokenHash, user.ID)
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "invite could not be accepted")
		return
	}
	_, sessionToken, err := s.auth.Login(r.Context(), invite.Email, req.Password)
	if err == nil {
		s.auth.SetSessionCookie(w, sessionToken)
	}
	s.sendVerificationEmail(r, user)
	httpx.JSON(w, http.StatusOK, accepted)
}

func (s *Server) updateMember(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	targetUserID := r.PathValue("userID")
	var req struct {
		Role string `json:"role"`
	}
	if err := httpx.Decode(r, &req); err != nil || req.Role == "" {
		httpx.Error(w, http.StatusBadRequest, "role is required")
		return
	}
	if _, ok := s.store.GetRole(practiceID, req.Role); !ok {
		httpx.Error(w, http.StatusBadRequest, "unknown role")
		return
	}
	member, ok := s.store.UpdateMemberRole(practiceID, targetUserID, req.Role)
	if !ok {
		httpx.Error(w, http.StatusNotFound, "member not found")
		return
	}
	s.audit(practiceID, "user", userID, "member.role_updated", "practice_member", targetUserID, map[string]string{"role": member.Role})
	httpx.JSON(w, http.StatusOK, member)
}

func (s *Server) disableMember(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	targetUserID := r.PathValue("userID")
	if targetUserID == userID {
		httpx.Error(w, http.StatusBadRequest, "cannot disable your own membership")
		return
	}
	if !s.store.DisableMember(practiceID, targetUserID) {
		httpx.Error(w, http.StatusNotFound, "member not found")
		return
	}
	s.audit(practiceID, "user", userID, "member.disabled", "practice_member", targetUserID, nil)
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "disabled"})
}

func (s *Server) voiceBootstrap(w http.ResponseWriter, r *http.Request) {
	if !s.verifyVoice(w, r) {
		return
	}
	var req struct {
		PracticeID string `json:"practiceId"`
	}
	if err := httpx.Decode(r, &req); err != nil || req.PracticeID == "" {
		httpx.Error(w, http.StatusBadRequest, "practiceId is required")
		return
	}
	bootstrap, ok := s.voice.Bootstrap(r.Context(), req.PracticeID)
	if !ok {
		httpx.Error(w, http.StatusNotFound, "practice not found")
		return
	}
	httpx.JSON(w, http.StatusOK, bootstrap)
}

func (s *Server) voicePracticeInfo(w http.ResponseWriter, r *http.Request) {
	if !s.verifyVoice(w, r) {
		return
	}
	var req struct {
		PracticeID string `json:"practiceId"`
	}
	if err := httpx.Decode(r, &req); err != nil || req.PracticeID == "" {
		httpx.Error(w, http.StatusBadRequest, "practiceId is required")
		return
	}
	config, ok := s.voice.GetPracticeInfo(r.Context(), req.PracticeID)
	if !ok {
		httpx.Error(w, http.StatusNotFound, "practice not found")
		return
	}
	httpx.JSON(w, http.StatusOK, config)
}

func (s *Server) voiceAppointmentRequest(w http.ResponseWriter, r *http.Request) {
	if !s.verifyVoice(w, r) {
		return
	}
	var req domain.AppointmentRequest
	if err := httpx.Decode(r, &req); err != nil || req.PracticeID == "" {
		httpx.Error(w, http.StatusBadRequest, "invalid appointment request")
		return
	}
	appointment := s.voice.CreateAppointmentRequest(r.Context(), req)
	if config, ok := s.store.GetAssistantConfig(req.PracticeID); ok {
		if err := s.notifications.SendAppointmentRequestEmail(r.Context(), config.NotificationEmail, appointment); err != nil {
			s.logger.Error("appointment notification failed", "practice_id", req.PracticeID, "error", err)
		}
	}
	s.audit(req.PracticeID, "webhook", "voice", "appointment_request.created", "appointment_request", appointment.ID, map[string]string{"callerName": appointment.CallerName, "requestType": appointment.RequestType})
	httpx.JSON(w, http.StatusCreated, appointment)
}

func (s *Server) voiceCallSummary(w http.ResponseWriter, r *http.Request) {
	if !s.verifyVoice(w, r) {
		return
	}
	var req domain.CallSummary
	if err := httpx.Decode(r, &req); err != nil || req.PracticeID == "" {
		httpx.Error(w, http.StatusBadRequest, "invalid call summary")
		return
	}
	summary := s.voice.CreateCallSummary(r.Context(), req)
	if config, ok := s.store.GetAssistantConfig(req.PracticeID); ok {
		if err := s.notifications.SendCallSummaryEmail(r.Context(), config.NotificationEmail, summary); err != nil {
			s.logger.Error("call summary notification failed", "practice_id", req.PracticeID, "error", err)
		}
	}
	s.audit(req.PracticeID, "webhook", "voice", "call_summary.created", "call_summary", summary.ID, map[string]string{"callerName": summary.CallerName, "urgency": summary.Urgency})
	httpx.JSON(w, http.StatusCreated, summary)
}

func (s *Server) verifyVoice(w http.ResponseWriter, r *http.Request) bool {
	secret := r.Header.Get("X-DentalDesk-Webhook-Secret")
	if err := s.voice.Verify(secret); err != nil {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	return true
}

func (s *Server) requireUser(next func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := s.auth.UserIDFromRequest(r)
		if !ok {
			httpx.Error(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next(w, r, userID)
	}
}

func (s *Server) requirePractice(next func(http.ResponseWriter, *http.Request, string, string)) http.HandlerFunc {
	return s.requireUser(func(w http.ResponseWriter, r *http.Request, userID string) {
		practiceID := r.PathValue("practiceID")
		if practiceID == "" || !s.practices.IsMember(r.Context(), practiceID, userID) {
			httpx.Error(w, http.StatusForbidden, "practice access denied")
			return
		}
		next(w, r, userID, practiceID)
	})
}

func (s *Server) requirePermission(permission string, next func(http.ResponseWriter, *http.Request, string, string)) http.HandlerFunc {
	return s.requireUser(func(w http.ResponseWriter, r *http.Request, userID string) {
		practiceID := r.PathValue("practiceID")
		if practiceID == "" || !s.store.HasPermission(practiceID, userID, permission) {
			httpx.Error(w, http.StatusForbidden, "permission denied")
			return
		}
		next(w, r, userID, practiceID)
	})
}

func normalizeProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "retell":
		return "retell"
	case "custom":
		return "custom"
	default:
		return "vapi"
	}
}

func webhookStatus(config domain.VoiceProviderConfig) string {
	if config.AssistantID == "" && config.PhoneNumber == "" {
		return "not_configured"
	}
	if config.AssistantID == "" || config.PhoneNumber == "" {
		return "needs_attention"
	}
	return "configured"
}

func normalizeCalendarMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "booking_link":
		return "booking_link"
	case "google":
		return "google"
	default:
		return "request_only"
	}
}

func normalizeCalendarProvider(provider, mode string) string {
	if mode == "google" {
		return "google"
	}
	if mode == "booking_link" {
		switch strings.ToLower(strings.TrimSpace(provider)) {
		case "calendly":
			return "calendly"
		case "google":
			return "google"
		case "custom":
			return "custom"
		default:
			return "booking_link"
		}
	}
	return "none"
}

func calendarStatus(config domain.CalendarConfig) string {
	if config.OAuthConnected {
		return "connected"
	}
	switch config.Mode {
	case "booking_link":
		if strings.TrimSpace(config.BookingURL) == "" {
			return "needs_booking_url"
		}
		return "configured"
	case "google":
		if strings.TrimSpace(config.CalendarID) == "" {
			return "needs_calendar_id"
		}
		return "ready_for_oauth"
	default:
		return "not_configured"
	}
}

func normalizeBillingPlan(plan string) string {
	switch strings.ToLower(strings.TrimSpace(plan)) {
	case "starter":
		return "starter"
	case "growth":
		return "growth"
	case "custom":
		return "custom"
	default:
		return "pilot"
	}
}

func normalizeBillingStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "trialing":
		return "trialing"
	case "active":
		return "active"
	case "past_due":
		return "past_due"
	case "canceled":
		return "canceled"
	default:
		return "manual"
	}
}

func validAppointmentStatus(status string) bool {
	switch status {
	case "new", "contacted", "scheduled", "closed", "spam":
		return true
	default:
		return false
	}
}

func (s *Server) audit(practiceID, actorType, actorID, action, targetType, targetID string, metadata map[string]string) {
	s.store.CreateAuditLog(domain.AuditLog{
		PracticeID: practiceID,
		ActorType:  actorType,
		ActorID:    actorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Metadata:   metadata,
	})
}

func (s *Server) allowAuthAttempt(r *http.Request, scope string) bool {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	if comma := strings.Index(ip, ","); comma >= 0 {
		ip = strings.TrimSpace(ip[:comma])
	}
	return s.authLimiter.Allow(scope + ":" + ip)
}

func (s *Server) sendVerificationEmail(r *http.Request, user domain.User) {
	token, err := s.auth.CreateEmailVerification(r.Context(), user.ID)
	if err != nil {
		s.logger.Error("email verification token failed", "user_id", user.ID, "error", err)
		return
	}
	verifyURL := s.cfg.WebBaseURL + "/verify-email?token=" + token
	if err := s.notifications.SendEmailVerification(r.Context(), user.Email, verifyURL); err != nil {
		s.logger.Error("email verification send failed", "user_id", user.ID, "error", err)
	}
}

func (s *Server) withCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/v1/voice/") {
			next.ServeHTTP(w, r)
			return
		}
		if _, err := r.Cookie(auth.SessionCookieName); err != nil {
			next.ServeHTTP(w, r)
			return
		}
		cookie, err := r.Cookie(auth.CSRFCookieName)
		if err != nil || cookie.Value == "" || r.Header.Get("X-CSRF-Token") != cookie.Value {
			httpx.Error(w, http.StatusForbidden, "csrf token required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == s.cfg.WebBaseURL || strings.HasPrefix(origin, "http://localhost:") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-DentalDesk-Webhook-Secret, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
