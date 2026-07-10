package domain

import "time"

type Practice struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Specialty string    `json:"specialty"`
	CreatedAt time.Time `json:"createdAt"`
}

type User struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	DisplayName     string     `json:"displayName"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

type PracticeMember struct {
	PracticeID  string `json:"practiceId"`
	UserID      string `json:"userId"`
	Role        string `json:"role"`
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Active      bool   `json:"active"`
}

type Role struct {
	ID          string    `json:"id"`
	PracticeID  string    `json:"practiceId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	System      bool      `json:"system"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"createdAt"`
}

type PracticeInvite struct {
	ID         string     `json:"id"`
	PracticeID string     `json:"practiceId"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	InvitedBy  string     `json:"invitedBy"`
	AcceptedAt *time.Time `json:"acceptedAt,omitempty"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type Location struct {
	ID           string    `json:"id"`
	PracticeID   string    `json:"practiceId"`
	Name         string    `json:"name"`
	AddressLine1 string    `json:"addressLine1"`
	AddressLine2 string    `json:"addressLine2"`
	City         string    `json:"city"`
	Region       string    `json:"region"`
	PostalCode   string    `json:"postalCode"`
	Country      string    `json:"country"`
	Timezone     string    `json:"timezone"`
	CreatedAt    time.Time `json:"createdAt"`
}

type AssistantConfig struct {
	ID                string            `json:"id"`
	PracticeID        string            `json:"practiceId"`
	Greeting          string            `json:"greeting"`
	EscalationPhone   string            `json:"escalationPhone"`
	NotificationEmail string            `json:"notificationEmail"`
	Settings          map[string]string `json:"settings"`
	CreatedAt         time.Time         `json:"createdAt"`
	UpdatedAt         time.Time         `json:"updatedAt"`
}

type VoiceProviderConfig struct {
	ID            string     `json:"id"`
	PracticeID    string     `json:"practiceId"`
	Provider      string     `json:"provider"`
	PhoneNumber   string     `json:"phoneNumber"`
	AssistantID   string     `json:"assistantId"`
	WebhookStatus string     `json:"webhookStatus"`
	LastWebhookAt *time.Time `json:"lastWebhookAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type CalendarConfig struct {
	ID                   string     `json:"id"`
	PracticeID           string     `json:"practiceId"`
	Mode                 string     `json:"mode"`
	Provider             string     `json:"provider"`
	BookingURL           string     `json:"bookingUrl"`
	CalendarID           string     `json:"calendarId"`
	Timezone             string     `json:"timezone"`
	Status               string     `json:"status"`
	Instructions         string     `json:"instructions"`
	OAuthConnected       bool       `json:"oauthConnected"`
	OAuthTokenExpiresAt  *time.Time `json:"oauthTokenExpiresAt,omitempty"`
	OAuthAccessTokenEnc  string     `json:"-"`
	OAuthRefreshTokenEnc string     `json:"-"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type BillingSubscription struct {
	ID                   string     `json:"id"`
	PracticeID           string     `json:"practiceId"`
	Plan                 string     `json:"plan"`
	Status               string     `json:"status"`
	IncludedMinutes      int        `json:"includedMinutes"`
	OverageCents         int        `json:"overageCents"`
	StripeCustomerID     string     `json:"stripeCustomerId"`
	StripeSubscriptionID string     `json:"stripeSubscriptionId"`
	TrialEndsAt          *time.Time `json:"trialEndsAt,omitempty"`
	CurrentPeriodEndsAt  *time.Time `json:"currentPeriodEndsAt,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type CallSession struct {
	ID             string     `json:"id"`
	PracticeID     string     `json:"practiceId"`
	Provider       string     `json:"provider"`
	ProviderCallID string     `json:"providerCallId"`
	CallerPhone    string     `json:"callerPhone"`
	Status         string     `json:"status"`
	StartedAt      time.Time  `json:"startedAt"`
	EndedAt        *time.Time `json:"endedAt,omitempty"`
}

type CallSummary struct {
	ID             string    `json:"id"`
	CallSessionID  string    `json:"callSessionId"`
	PracticeID     string    `json:"practiceId"`
	CallerName     string    `json:"callerName"`
	Reason         string    `json:"reason"`
	Urgency        string    `json:"urgency"`
	AIAction       string    `json:"aiAction"`
	FollowUpNeeded string    `json:"followUpNeeded"`
	Summary        string    `json:"summary"`
	CreatedAt      time.Time `json:"createdAt"`
}

type AppointmentRequest struct {
	ID            string    `json:"id"`
	PracticeID    string    `json:"practiceId"`
	CallSessionID string    `json:"callSessionId,omitempty"`
	CallerName    string    `json:"callerName"`
	CallerPhone   string    `json:"callerPhone"`
	RequestType   string    `json:"requestType"`
	PreferredTime string    `json:"preferredTime"`
	Insurance     string    `json:"insurance"`
	Notes         string    `json:"notes"`
	StaffNote     string    `json:"staffNote"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
}

type AuditLog struct {
	ID         string            `json:"id"`
	PracticeID string            `json:"practiceId,omitempty"`
	ActorType  string            `json:"actorType"`
	ActorID    string            `json:"actorId"`
	Action     string            `json:"action"`
	TargetType string            `json:"targetType"`
	TargetID   string            `json:"targetId"`
	Metadata   map[string]string `json:"metadata"`
	CreatedAt  time.Time         `json:"createdAt"`
}
