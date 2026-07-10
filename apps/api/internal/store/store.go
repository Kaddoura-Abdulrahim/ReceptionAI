package store

import (
	"time"

	"dentaldesk/apps/api/internal/domain"
)

type Store interface {
	CreateInvite(practiceID, email, role, invitedBy, tokenHash string, expiresAt time.Time) domain.PracticeInvite
	ListInvites(practiceID string) []domain.PracticeInvite
	GetInviteByTokenHash(tokenHash string) (domain.PracticeInvite, bool)
	AcceptInvite(tokenHash, userID string) (domain.PracticeInvite, bool)

	CreateUser(email, displayName, passwordHash string) domain.User
	FindUserByEmail(email string) (domain.User, string, bool)
	CreateEmailVerificationToken(userID, tokenHash string, expiresAt time.Time)
	VerifyEmailByTokenHash(tokenHash string) (domain.User, bool)
	CreatePasswordResetToken(email, tokenHash string, expiresAt time.Time) (domain.User, bool)
	ResetPasswordByTokenHash(tokenHash, passwordHash string) (domain.User, bool)
	CreateSession(userID, tokenHash string, expiresAt time.Time) Session
	FindSession(tokenHash string) (Session, bool)

	CreatePractice(name, ownerUserID string) domain.Practice
	ListPracticesForUser(userID string) []domain.Practice
	IsMember(practiceID, userID string) bool
	GetMember(practiceID, userID string) (domain.PracticeMember, bool)
	ListMembers(practiceID string) []domain.PracticeMember
	UpsertMember(practiceID, userID, role string) domain.PracticeMember
	UpdateMemberRole(practiceID, userID, role string) (domain.PracticeMember, bool)
	DisableMember(practiceID, userID string) bool

	ListRoles(practiceID string) []domain.Role
	GetRole(practiceID, name string) (domain.Role, bool)
	CreateRole(practiceID, name, description string, permissions []string) domain.Role

	CreateLocation(location domain.Location) domain.Location
	ListLocations(practiceID string) []domain.Location
	HasPermission(practiceID, userID, permission string) bool

	GetAssistantConfig(practiceID string) (domain.AssistantConfig, bool)
	SaveAssistantConfig(config domain.AssistantConfig) domain.AssistantConfig
	GetVoiceProviderConfig(practiceID string) (domain.VoiceProviderConfig, bool)
	SaveVoiceProviderConfig(config domain.VoiceProviderConfig) domain.VoiceProviderConfig
	GetCalendarConfig(practiceID string) (domain.CalendarConfig, bool)
	SaveCalendarConfig(config domain.CalendarConfig) domain.CalendarConfig
	SaveCalendarOAuth(practiceID, accessTokenEnc, refreshTokenEnc string, expiresAt time.Time) (domain.CalendarConfig, bool)
	GetBillingSubscription(practiceID string) (domain.BillingSubscription, bool)
	SaveBillingSubscription(subscription domain.BillingSubscription) domain.BillingSubscription
	CreateCall(call domain.CallSession) domain.CallSession
	CreateSummary(summary domain.CallSummary) domain.CallSummary
	ListSummaries(practiceID string) []domain.CallSummary
	CreateAppointment(request domain.AppointmentRequest) domain.AppointmentRequest
	ListAppointments(practiceID string) []domain.AppointmentRequest
	UpdateAppointmentRequest(practiceID, requestID, status, staffNote string) (domain.AppointmentRequest, bool)
	CreateAuditLog(log domain.AuditLog) domain.AuditLog
	ListAuditLogs(practiceID string, limit int) []domain.AuditLog
}

var _ Store = (*MemoryStore)(nil)
