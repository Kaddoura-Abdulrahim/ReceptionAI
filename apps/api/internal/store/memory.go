package store

import (
	"errors"
	"strings"
	"sync"
	"time"

	"dentaldesk/apps/api/internal/domain"
	"dentaldesk/apps/api/internal/platform/id"
)

var ErrNotFound = errors.New("not found")

type MemoryStore struct {
	mu                   sync.RWMutex
	usersByEmail         map[string]domain.User
	passwordHashByUser   map[string]string
	sessionsByTokenHash  map[string]Session
	emailVerifyByToken   map[string]userToken
	passwordResetByToken map[string]userToken
	practices            map[string]domain.Practice
	members              map[string]domain.PracticeMember
	rolesByPractice      map[string][]domain.Role
	locationsByPractice  map[string][]domain.Location
	invitesByTokenHash   map[string]domain.PracticeInvite
	invitesByPractice    map[string][]string
	configsByPractice    map[string]domain.AssistantConfig
	voiceByPractice      map[string]domain.VoiceProviderConfig
	calendarByPractice   map[string]domain.CalendarConfig
	billingByPractice    map[string]domain.BillingSubscription
	calls                map[string]domain.CallSession
	summaries            []domain.CallSummary
	appointments         []domain.AppointmentRequest
	auditLogs            []domain.AuditLog
}

type Session struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
}

type userToken struct {
	UserID    string
	ExpiresAt time.Time
	UsedAt    *time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		usersByEmail:         make(map[string]domain.User),
		passwordHashByUser:   make(map[string]string),
		sessionsByTokenHash:  make(map[string]Session),
		emailVerifyByToken:   make(map[string]userToken),
		passwordResetByToken: make(map[string]userToken),
		practices:            make(map[string]domain.Practice),
		members:              make(map[string]domain.PracticeMember),
		rolesByPractice:      make(map[string][]domain.Role),
		locationsByPractice:  make(map[string][]domain.Location),
		invitesByTokenHash:   make(map[string]domain.PracticeInvite),
		invitesByPractice:    make(map[string][]string),
		configsByPractice:    make(map[string]domain.AssistantConfig),
		voiceByPractice:      make(map[string]domain.VoiceProviderConfig),
		calendarByPractice:   make(map[string]domain.CalendarConfig),
		billingByPractice:    make(map[string]domain.BillingSubscription),
		calls:                make(map[string]domain.CallSession),
	}
}

func (s *MemoryStore) CreateInvite(practiceID, email, role, invitedBy, tokenHash string, expiresAt time.Time) domain.PracticeInvite {
	s.mu.Lock()
	defer s.mu.Unlock()
	invite := domain.PracticeInvite{
		ID:         id.New(),
		PracticeID: practiceID,
		Email:      strings.ToLower(strings.TrimSpace(email)),
		Role:       strings.ToLower(strings.TrimSpace(role)),
		InvitedBy:  invitedBy,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now().UTC(),
	}
	s.invitesByTokenHash[tokenHash] = invite
	s.invitesByPractice[practiceID] = append(s.invitesByPractice[practiceID], tokenHash)
	return invite
}

func (s *MemoryStore) ListInvites(practiceID string) []domain.PracticeInvite {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.PracticeInvite, 0)
	for _, tokenHash := range s.invitesByPractice[practiceID] {
		if invite, ok := s.invitesByTokenHash[tokenHash]; ok {
			out = append(out, invite)
		}
	}
	return out
}

func (s *MemoryStore) GetInviteByTokenHash(tokenHash string) (domain.PracticeInvite, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	invite, ok := s.invitesByTokenHash[tokenHash]
	if !ok || invite.AcceptedAt != nil || time.Now().UTC().After(invite.ExpiresAt) {
		return domain.PracticeInvite{}, false
	}
	return invite, true
}

func (s *MemoryStore) AcceptInvite(tokenHash, userID string) (domain.PracticeInvite, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	invite, ok := s.invitesByTokenHash[tokenHash]
	if !ok || invite.AcceptedAt != nil || time.Now().UTC().After(invite.ExpiresAt) {
		return domain.PracticeInvite{}, false
	}
	now := time.Now().UTC()
	invite.AcceptedAt = &now
	s.invitesByTokenHash[tokenHash] = invite
	s.members[invite.PracticeID+":"+userID] = domain.PracticeMember{
		PracticeID: invite.PracticeID,
		UserID:     userID,
		Role:       invite.Role,
		Active:     true,
	}
	return invite, true
}

func (s *MemoryStore) CreateUser(email, displayName, passwordHash string) domain.User {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := domain.User{
		ID:          id.New(),
		Email:       strings.ToLower(strings.TrimSpace(email)),
		DisplayName: strings.TrimSpace(displayName),
		CreatedAt:   time.Now().UTC(),
	}
	s.usersByEmail[user.Email] = user
	s.passwordHashByUser[user.ID] = passwordHash
	return user
}

func (s *MemoryStore) FindUserByEmail(email string) (domain.User, string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.usersByEmail[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return domain.User{}, "", false
	}
	return user, s.passwordHashByUser[user.ID], true
}

func (s *MemoryStore) CreateEmailVerificationToken(userID, tokenHash string, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emailVerifyByToken[tokenHash] = userToken{UserID: userID, ExpiresAt: expiresAt}
}

func (s *MemoryStore) VerifyEmailByTokenHash(tokenHash string) (domain.User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	token, ok := s.emailVerifyByToken[tokenHash]
	if !ok || token.UsedAt != nil || time.Now().UTC().After(token.ExpiresAt) {
		return domain.User{}, false
	}
	now := time.Now().UTC()
	token.UsedAt = &now
	s.emailVerifyByToken[tokenHash] = token
	for email, user := range s.usersByEmail {
		if user.ID == token.UserID {
			user.EmailVerifiedAt = &now
			s.usersByEmail[email] = user
			return user, true
		}
	}
	return domain.User{}, false
}

func (s *MemoryStore) CreatePasswordResetToken(email, tokenHash string, expiresAt time.Time) (domain.User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.usersByEmail[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return domain.User{}, false
	}
	s.passwordResetByToken[tokenHash] = userToken{UserID: user.ID, ExpiresAt: expiresAt}
	return user, true
}

func (s *MemoryStore) ResetPasswordByTokenHash(tokenHash, passwordHash string) (domain.User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	token, ok := s.passwordResetByToken[tokenHash]
	if !ok || token.UsedAt != nil || time.Now().UTC().After(token.ExpiresAt) {
		return domain.User{}, false
	}
	now := time.Now().UTC()
	token.UsedAt = &now
	s.passwordResetByToken[tokenHash] = token
	for _, user := range s.usersByEmail {
		if user.ID == token.UserID {
			s.passwordHashByUser[user.ID] = passwordHash
			return user, true
		}
	}
	return domain.User{}, false
}

func (s *MemoryStore) CreateSession(userID, tokenHash string, expiresAt time.Time) Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	session := Session{ID: id.New(), UserID: userID, TokenHash: tokenHash, ExpiresAt: expiresAt}
	s.sessionsByTokenHash[tokenHash] = session
	return session
}

func (s *MemoryStore) FindSession(tokenHash string) (Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessionsByTokenHash[tokenHash]
	return session, ok && time.Now().UTC().Before(session.ExpiresAt)
}

func (s *MemoryStore) CreatePractice(name, ownerUserID string) domain.Practice {
	s.mu.Lock()
	defer s.mu.Unlock()
	practice := domain.Practice{ID: id.New(), Name: strings.TrimSpace(name), Specialty: "dental", CreatedAt: time.Now().UTC()}
	s.practices[practice.ID] = practice
	s.members[practice.ID+":"+ownerUserID] = domain.PracticeMember{PracticeID: practice.ID, UserID: ownerUserID, Role: "owner", Active: true}
	s.rolesByPractice[practice.ID] = defaultRoles(practice.ID)
	s.configsByPractice[practice.ID] = domain.AssistantConfig{
		ID:                id.New(),
		PracticeID:        practice.ID,
		Greeting:          "Thank you for calling " + practice.Name + ". How can I help you today?",
		EscalationPhone:   "",
		NotificationEmail: "",
		Settings:          map[string]string{"specialty": "dental"},
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	s.voiceByPractice[practice.ID] = domain.VoiceProviderConfig{
		ID:            id.New(),
		PracticeID:    practice.ID,
		Provider:      "vapi",
		WebhookStatus: "not_configured",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	s.calendarByPractice[practice.ID] = domain.CalendarConfig{
		ID:         id.New(),
		PracticeID: practice.ID,
		Mode:       "request_only",
		Provider:   "none",
		Timezone:   "America/New_York",
		Status:     "not_configured",
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	s.billingByPractice[practice.ID] = defaultBillingSubscription(practice.ID)
	return practice
}

func defaultRoles(practiceID string) []domain.Role {
	now := time.Now().UTC()
	return []domain.Role{
		{ID: id.New(), PracticeID: practiceID, Name: "owner", Description: "Full practice access", System: true, Permissions: AllPermissions(), CreatedAt: now},
		{ID: id.New(), PracticeID: practiceID, Name: "admin", Description: "Manage practice operations", System: true, Permissions: AdminPermissions(), CreatedAt: now},
		{ID: id.New(), PracticeID: practiceID, Name: "staff", Description: "Handle calls and appointment requests", System: true, Permissions: StaffPermissions(), CreatedAt: now},
		{ID: id.New(), PracticeID: practiceID, Name: "billing", Description: "Manage billing", System: true, Permissions: []string{"billing:read", "billing:update"}, CreatedAt: now},
		{ID: id.New(), PracticeID: practiceID, Name: "viewer", Description: "Read-only access", System: true, Permissions: []string{"call:read", "appointment:read", "assistant_config:read"}, CreatedAt: now},
	}
}

func AllPermissions() []string {
	return []string{
		"practice:create", "practice:update", "practice:delete",
		"location:create", "location:update", "location:delete",
		"member:invite", "member:update", "member:disable",
		"role:create", "role:update", "role:delete",
		"call:read", "call:update",
		"appointment:read", "appointment:update",
		"assistant_config:read", "assistant_config:update",
		"calendar:read", "calendar:update",
		"billing:read", "billing:update",
	}
}

func AdminPermissions() []string {
	return []string{
		"practice:update",
		"location:create", "location:update", "location:delete",
		"member:invite", "member:update", "member:disable",
		"role:create", "role:update", "role:delete",
		"call:read", "call:update",
		"appointment:read", "appointment:update",
		"assistant_config:read", "assistant_config:update",
		"calendar:read", "calendar:update",
	}
}

func StaffPermissions() []string {
	return []string{
		"call:read", "call:update",
		"appointment:read", "appointment:update",
		"assistant_config:read", "calendar:read",
	}
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

func defaultBillingSubscription(practiceID string) domain.BillingSubscription {
	now := time.Now().UTC()
	trialEndsAt := now.AddDate(0, 0, 14)
	return domain.BillingSubscription{
		ID:              id.New(),
		PracticeID:      practiceID,
		Plan:            "pilot",
		Status:          "manual",
		IncludedMinutes: 300,
		OverageCents:    25,
		TrialEndsAt:     &trialEndsAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (s *MemoryStore) ListPracticesForUser(userID string) []domain.Practice {
	s.mu.RLock()
	defer s.mu.RUnlock()
	practices := make([]domain.Practice, 0)
	for _, member := range s.members {
		if member.UserID == userID {
			if practice, ok := s.practices[member.PracticeID]; ok {
				practices = append(practices, practice)
			}
		}
	}
	return practices
}

func (s *MemoryStore) IsMember(practiceID, userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	member, ok := s.members[practiceID+":"+userID]
	return ok && member.Active
}

func (s *MemoryStore) GetMember(practiceID, userID string) (domain.PracticeMember, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	member, ok := s.members[practiceID+":"+userID]
	if !ok || !member.Active {
		return domain.PracticeMember{}, false
	}
	if user, _, ok := s.findUserByIDLocked(userID); ok {
		member.Email = user.Email
		member.DisplayName = user.DisplayName
	}
	return member, true
}

func (s *MemoryStore) ListMembers(practiceID string) []domain.PracticeMember {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.PracticeMember, 0)
	for _, member := range s.members {
		if member.PracticeID == practiceID && member.Active {
			if user, _, ok := s.findUserByIDLocked(member.UserID); ok {
				member.Email = user.Email
				member.DisplayName = user.DisplayName
			}
			out = append(out, member)
		}
	}
	return out
}

func (s *MemoryStore) UpsertMember(practiceID, userID, role string) domain.PracticeMember {
	s.mu.Lock()
	defer s.mu.Unlock()
	member := domain.PracticeMember{PracticeID: practiceID, UserID: userID, Role: role, Active: true}
	s.members[practiceID+":"+userID] = member
	return member
}

func (s *MemoryStore) UpdateMemberRole(practiceID, userID, role string) (domain.PracticeMember, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := practiceID + ":" + userID
	member, ok := s.members[key]
	if !ok {
		return domain.PracticeMember{}, false
	}
	member.Role = role
	s.members[key] = member
	return member, true
}

func (s *MemoryStore) DisableMember(practiceID, userID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := practiceID + ":" + userID
	member, ok := s.members[key]
	if !ok {
		return false
	}
	member.Active = false
	s.members[key] = member
	return true
}

func (s *MemoryStore) ListRoles(practiceID string) []domain.Role {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.Role(nil), s.rolesByPractice[practiceID]...)
}

func (s *MemoryStore) GetRole(practiceID, name string) (domain.Role, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getRoleLocked(practiceID, name)
}

func (s *MemoryStore) CreateRole(practiceID, name, description string, permissions []string) domain.Role {
	s.mu.Lock()
	defer s.mu.Unlock()
	role := domain.Role{
		ID:          id.New(),
		PracticeID:  practiceID,
		Name:        strings.ToLower(strings.TrimSpace(name)),
		Description: strings.TrimSpace(description),
		System:      false,
		Permissions: permissions,
		CreatedAt:   time.Now().UTC(),
	}
	s.rolesByPractice[practiceID] = append(s.rolesByPractice[practiceID], role)
	return role
}

func (s *MemoryStore) CreateLocation(location domain.Location) domain.Location {
	s.mu.Lock()
	defer s.mu.Unlock()
	if location.ID == "" {
		location.ID = id.New()
	}
	if location.Country == "" {
		location.Country = "US"
	}
	if location.Timezone == "" {
		location.Timezone = "America/New_York"
	}
	location.CreatedAt = time.Now().UTC()
	s.locationsByPractice[location.PracticeID] = append(s.locationsByPractice[location.PracticeID], location)
	return location
}

func (s *MemoryStore) ListLocations(practiceID string) []domain.Location {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.Location(nil), s.locationsByPractice[practiceID]...)
}

func (s *MemoryStore) HasPermission(practiceID, userID, permission string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	member, ok := s.members[practiceID+":"+userID]
	if !ok || !member.Active {
		return false
	}
	role, ok := s.getRoleLocked(practiceID, member.Role)
	if !ok {
		return false
	}
	for _, candidate := range role.Permissions {
		if candidate == permission {
			return true
		}
	}
	return false
}

func (s *MemoryStore) getRoleLocked(practiceID, name string) (domain.Role, bool) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	for _, role := range s.rolesByPractice[practiceID] {
		if role.Name == normalized {
			return role, true
		}
	}
	return domain.Role{}, false
}

func (s *MemoryStore) findUserByIDLocked(userID string) (domain.User, string, bool) {
	for _, user := range s.usersByEmail {
		if user.ID == userID {
			return user, s.passwordHashByUser[user.ID], true
		}
	}
	return domain.User{}, "", false
}

func (s *MemoryStore) GetAssistantConfig(practiceID string) (domain.AssistantConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	config, ok := s.configsByPractice[practiceID]
	return config, ok
}

func (s *MemoryStore) SaveAssistantConfig(config domain.AssistantConfig) domain.AssistantConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	config.UpdatedAt = time.Now().UTC()
	s.configsByPractice[config.PracticeID] = config
	return config
}

func (s *MemoryStore) GetVoiceProviderConfig(practiceID string) (domain.VoiceProviderConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	config, ok := s.voiceByPractice[practiceID]
	return config, ok
}

func (s *MemoryStore) SaveVoiceProviderConfig(config domain.VoiceProviderConfig) domain.VoiceProviderConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	if config.ID == "" {
		config.ID = id.New()
	}
	if config.Provider == "" {
		config.Provider = "vapi"
	}
	if config.WebhookStatus == "" {
		config.WebhookStatus = "not_configured"
	}
	now := time.Now().UTC()
	if config.CreatedAt.IsZero() {
		config.CreatedAt = now
	}
	config.UpdatedAt = now
	s.voiceByPractice[config.PracticeID] = config
	return config
}

func (s *MemoryStore) GetCalendarConfig(practiceID string) (domain.CalendarConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	config, ok := s.calendarByPractice[practiceID]
	return config, ok
}

func (s *MemoryStore) SaveCalendarConfig(config domain.CalendarConfig) domain.CalendarConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.calendarByPractice[config.PracticeID]
	if config.ID == "" {
		config.ID = current.ID
	}
	if config.ID == "" {
		config.ID = id.New()
	}
	if config.Mode == "" {
		config.Mode = "request_only"
	}
	if config.Provider == "" {
		config.Provider = "none"
	}
	if config.Timezone == "" {
		config.Timezone = "America/New_York"
	}
	config.Status = calendarStatus(config)
	config.OAuthConnected = current.OAuthConnected
	config.OAuthAccessTokenEnc = current.OAuthAccessTokenEnc
	config.OAuthRefreshTokenEnc = current.OAuthRefreshTokenEnc
	config.OAuthTokenExpiresAt = current.OAuthTokenExpiresAt
	if config.OAuthConnected {
		config.Status = "connected"
	}
	now := time.Now().UTC()
	if config.CreatedAt.IsZero() {
		config.CreatedAt = current.CreatedAt
	}
	if config.CreatedAt.IsZero() {
		config.CreatedAt = now
	}
	config.UpdatedAt = now
	s.calendarByPractice[config.PracticeID] = config
	return config
}

func (s *MemoryStore) SaveCalendarOAuth(practiceID, accessTokenEnc, refreshTokenEnc string, expiresAt time.Time) (domain.CalendarConfig, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	config, ok := s.calendarByPractice[practiceID]
	if !ok {
		return domain.CalendarConfig{}, false
	}
	config.Mode = "google"
	config.Provider = "google"
	config.Status = "connected"
	config.OAuthConnected = true
	config.OAuthAccessTokenEnc = accessTokenEnc
	if refreshTokenEnc != "" {
		config.OAuthRefreshTokenEnc = refreshTokenEnc
	}
	config.OAuthTokenExpiresAt = &expiresAt
	config.UpdatedAt = time.Now().UTC()
	s.calendarByPractice[practiceID] = config
	return config, true
}

func (s *MemoryStore) GetBillingSubscription(practiceID string) (domain.BillingSubscription, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	subscription, ok := s.billingByPractice[practiceID]
	return subscription, ok
}

func (s *MemoryStore) SaveBillingSubscription(subscription domain.BillingSubscription) domain.BillingSubscription {
	s.mu.Lock()
	defer s.mu.Unlock()
	if subscription.ID == "" {
		subscription.ID = id.New()
	}
	if subscription.Plan == "" {
		subscription.Plan = "pilot"
	}
	if subscription.Status == "" {
		subscription.Status = "manual"
	}
	if subscription.IncludedMinutes <= 0 {
		subscription.IncludedMinutes = 300
	}
	if subscription.OverageCents <= 0 {
		subscription.OverageCents = 25
	}
	now := time.Now().UTC()
	if subscription.CreatedAt.IsZero() {
		subscription.CreatedAt = now
	}
	subscription.UpdatedAt = now
	s.billingByPractice[subscription.PracticeID] = subscription
	return subscription
}

func (s *MemoryStore) CreateCall(call domain.CallSession) domain.CallSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	if call.ID == "" {
		call.ID = id.New()
	}
	if call.StartedAt.IsZero() {
		call.StartedAt = time.Now().UTC()
	}
	s.calls[call.ID] = call
	return call
}

func (s *MemoryStore) CreateSummary(summary domain.CallSummary) domain.CallSummary {
	s.mu.Lock()
	defer s.mu.Unlock()
	if summary.ID == "" {
		summary.ID = id.New()
	}
	if summary.CreatedAt.IsZero() {
		summary.CreatedAt = time.Now().UTC()
	}
	s.summaries = append(s.summaries, summary)
	return summary
}

func (s *MemoryStore) ListSummaries(practiceID string) []domain.CallSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.CallSummary, 0)
	for _, summary := range s.summaries {
		if summary.PracticeID == practiceID {
			out = append(out, summary)
		}
	}
	return out
}

func (s *MemoryStore) CreateAppointment(request domain.AppointmentRequest) domain.AppointmentRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	if request.ID == "" {
		request.ID = id.New()
	}
	if request.CreatedAt.IsZero() {
		request.CreatedAt = time.Now().UTC()
	}
	if request.Status == "" {
		request.Status = "new"
	}
	s.appointments = append(s.appointments, request)
	return request
}

func (s *MemoryStore) ListAppointments(practiceID string) []domain.AppointmentRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.AppointmentRequest, 0)
	for _, request := range s.appointments {
		if request.PracticeID == practiceID {
			out = append(out, request)
		}
	}
	return out
}

func (s *MemoryStore) UpdateAppointmentRequest(practiceID, requestID, status, staffNote string) (domain.AppointmentRequest, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, request := range s.appointments {
		if request.PracticeID == practiceID && request.ID == requestID {
			if status != "" {
				request.Status = status
			}
			request.StaffNote = staffNote
			s.appointments[i] = request
			return request, true
		}
	}
	return domain.AppointmentRequest{}, false
}

func (s *MemoryStore) CreateAuditLog(log domain.AuditLog) domain.AuditLog {
	s.mu.Lock()
	defer s.mu.Unlock()
	if log.ID == "" {
		log.ID = id.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}
	if log.Metadata == nil {
		log.Metadata = map[string]string{}
	}
	s.auditLogs = append(s.auditLogs, log)
	return log
}

func (s *MemoryStore) ListAuditLogs(practiceID string, limit int) []domain.AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	out := make([]domain.AuditLog, 0, limit)
	for i := len(s.auditLogs) - 1; i >= 0 && len(out) < limit; i-- {
		log := s.auditLogs[i]
		if log.PracticeID == practiceID {
			out = append(out, log)
		}
	}
	return out
}
