package voice

import (
	"context"
	"errors"
	"strings"

	"dentaldesk/apps/api/internal/domain"
	"dentaldesk/apps/api/internal/store"
)

var ErrUnauthorized = errors.New("unauthorized voice request")

type Service struct {
	store         store.Store
	webhookSecret string
}

type Bootstrap struct {
	PracticeID    string       `json:"practiceId"`
	SystemPrompt  string       `json:"systemPrompt"`
	FirstMessage  string       `json:"firstMessage"`
	VoiceTone     string       `json:"voiceTone"`
	ToolEndpoints []ToolSchema `json:"toolEndpoints"`
}

type ToolSchema struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Parameters  map[string]string `json:"parameters"`
}

func NewService(store store.Store, webhookSecret string) *Service {
	return &Service{store: store, webhookSecret: webhookSecret}
}

func (s *Service) Verify(secret string) error {
	if s.webhookSecret == "" || secret != s.webhookSecret {
		return ErrUnauthorized
	}
	return nil
}

func (s *Service) GetPracticeInfo(ctx context.Context, practiceID string) (domain.AssistantConfig, bool) {
	_ = ctx
	return s.store.GetAssistantConfig(practiceID)
}

func (s *Service) Bootstrap(ctx context.Context, practiceID string) (Bootstrap, bool) {
	config, ok := s.GetPracticeInfo(ctx, practiceID)
	if !ok {
		return Bootstrap{}, false
	}
	calendar, _ := s.store.GetCalendarConfig(practiceID)
	return Bootstrap{
		PracticeID:    practiceID,
		SystemPrompt:  BuildDentalPrompt(config, calendar),
		FirstMessage:  config.Greeting,
		VoiceTone:     setting(config, "voiceTone", "warm, concise, professional"),
		ToolEndpoints: ToolSchemas(),
	}, true
}

func (s *Service) CreateAppointmentRequest(ctx context.Context, req domain.AppointmentRequest) domain.AppointmentRequest {
	_ = ctx
	req.RequestType = strings.TrimSpace(req.RequestType)
	return s.store.CreateAppointment(req)
}

func (s *Service) CreateCallSummary(ctx context.Context, summary domain.CallSummary) domain.CallSummary {
	_ = ctx
	if summary.Urgency == "" {
		summary.Urgency = "routine"
	}
	return s.store.CreateSummary(summary)
}

func BuildDentalPrompt(config domain.AssistantConfig, calendar domain.CalendarConfig) string {
	return strings.TrimSpace(`You are DentalDesk AI, an administrative phone receptionist for a dental practice.

Practice operating profile:
- Greeting: ` + config.Greeting + `
- Office hours: ` + setting(config, "officeHours", "Not configured") + `
- Services: ` + setting(config, "services", "Not configured") + `
- Accepted insurance: ` + setting(config, "acceptedInsurance", "Not configured") + `
- New patient rules: ` + setting(config, "newPatientRules", "Not configured") + `
- Cancellation policy: ` + setting(config, "cancellationPolicy", "Not configured") + `
- Intake form link: ` + setting(config, "intakeFormLink", "Not configured") + `
- Scheduling mode: ` + calendarModeText(calendar) + `
- Booking link: ` + emptyAs(calendar.BookingURL, "Not configured") + `
- Calendar instructions: ` + emptyAs(calendar.Instructions, "Not configured") + `
- Staff handoff phone: ` + emptyAs(config.EscalationPhone, "Not configured") + `
- Staff notification email: ` + emptyAs(config.NotificationEmail, "Not configured") + `
- Voice tone: ` + setting(config, "voiceTone", "warm, concise, professional") + `

Your job:
- Answer routine administrative questions.
- Collect name, phone number, caller type, reason for call, preferred time, insurance, and concise notes.
- Create appointment requests for new patient, existing patient, cleaning, exam, whitening, consultation, emergency dental visit, reschedule, and cancellation requests.
- If scheduling mode is booking-link and a booking link is configured, offer the booking link and still create an appointment request for staff visibility.
- If scheduling mode is request-only, do not promise a confirmed appointment time. Capture the request and tell the caller staff will confirm.
- Take messages when the caller needs staff follow-up.
- Keep responses brief and natural for a phone call.
- Confirm important details back to the caller.

Strict safety rules:
- Do not diagnose conditions.
- Do not recommend treatment.
- Do not give dental, medical, medication, prescription, or post-procedure clinical advice.
- Do not promise insurance coverage.
- Do not access or discuss patient records unless the workflow explicitly supports it.
- If the caller asks for advice, say you can help schedule or take a message but cannot provide dental advice.

Escalate immediately if the caller mentions severe pain, swelling, heavy bleeding, facial trauma, trouble breathing, post-surgery complication, medication question, prescription request, diagnosis request, treatment recommendation request, angry or distressed caller, billing dispute, request for a specific staff member, existing patient record details, uncertainty, or if the caller asks for a human.

Emergency language:
"I can help schedule an appointment or take a message for the dental team, but I cannot provide dental advice. If you are experiencing severe pain, swelling, heavy bleeding, trauma, or trouble breathing, seek urgent medical care or call emergency services."

At the end of each call, create a call summary with urgency, action taken, and follow-up needed.`)
}

func calendarModeText(calendar domain.CalendarConfig) string {
	switch calendar.Mode {
	case "booking_link":
		return "booking link"
	case "google":
		return "Google Calendar ready, but direct booking is not enabled until OAuth is connected"
	default:
		return "request only"
	}
}

func ToolSchemas() []ToolSchema {
	return []ToolSchema{
		{
			Name:        "getPracticeInfo",
			Description: "Fetch the configured dental practice profile and assistant rules.",
			Method:      "POST",
			Path:        "/v1/voice/practice-info",
			Parameters:  map[string]string{"practiceId": "string"},
		},
		{
			Name:        "createAppointmentRequest",
			Description: "Create a staff-review appointment request after collecting caller details.",
			Method:      "POST",
			Path:        "/v1/voice/appointment-request",
			Parameters: map[string]string{
				"practiceId":    "string",
				"callerName":    "string",
				"callerPhone":   "string",
				"requestType":   "string",
				"preferredTime": "string",
				"insurance":     "string",
				"notes":         "string",
			},
		},
		{
			Name:        "createCallSummary",
			Description: "Save the final call summary, urgency, action, and follow-up needed.",
			Method:      "POST",
			Path:        "/v1/voice/call-summary",
			Parameters: map[string]string{
				"practiceId":     "string",
				"callerName":     "string",
				"reason":         "string",
				"urgency":        "routine|urgent",
				"aiAction":       "string",
				"followUpNeeded": "string",
				"summary":        "string",
			},
		},
	}
}

func setting(config domain.AssistantConfig, key, fallback string) string {
	if config.Settings == nil {
		return fallback
	}
	value := strings.TrimSpace(config.Settings[key])
	if value == "" {
		return fallback
	}
	return value
}

func emptyAs(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
