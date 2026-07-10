package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"dentaldesk/apps/api/internal/domain"
	"dentaldesk/apps/api/internal/platform/id"
)

var _ Store = (*PostgresStore)(nil)

func (s *PostgresStore) CreateInvite(practiceID, email, role, invitedBy, tokenHash string, expiresAt time.Time) domain.PracticeInvite {
	invite := domain.PracticeInvite{
		ID:         id.New(),
		PracticeID: practiceID,
		Email:      strings.ToLower(strings.TrimSpace(email)),
		Role:       strings.ToLower(strings.TrimSpace(role)),
		InvitedBy:  invitedBy,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now().UTC(),
	}
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO practice_invites (id, practice_id, email, role, token_hash, invited_by_user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, invite.ID, invite.PracticeID, invite.Email, invite.Role, tokenHash, invite.InvitedBy, invite.ExpiresAt, invite.CreatedAt)
	if err != nil {
		panic(err)
	}
	return invite
}

func (s *PostgresStore) ListInvites(practiceID string) []domain.PracticeInvite {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id, practice_id, email, role, invited_by_user_id, accepted_at, expires_at, created_at
		FROM practice_invites
		WHERE practice_id = $1
		ORDER BY created_at DESC
	`, practiceID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	return scanInvites(rows)
}

func (s *PostgresStore) GetInviteByTokenHash(tokenHash string) (domain.PracticeInvite, bool) {
	var invite domain.PracticeInvite
	var acceptedAt sql.NullTime
	err := s.db.QueryRowContext(context.Background(), `
		SELECT id, practice_id, email, role, invited_by_user_id, accepted_at, expires_at, created_at
		FROM practice_invites
		WHERE token_hash = $1 AND accepted_at IS NULL AND expires_at > now()
	`, tokenHash).Scan(&invite.ID, &invite.PracticeID, &invite.Email, &invite.Role, &invite.InvitedBy, &acceptedAt, &invite.ExpiresAt, &invite.CreatedAt)
	if err != nil {
		return domain.PracticeInvite{}, false
	}
	if acceptedAt.Valid {
		invite.AcceptedAt = &acceptedAt.Time
	}
	return invite, true
}

func (s *PostgresStore) AcceptInvite(tokenHash, userID string) (domain.PracticeInvite, bool) {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	defer rollback(tx)

	var invite domain.PracticeInvite
	err = tx.QueryRowContext(context.Background(), `
		UPDATE practice_invites
		SET accepted_at = now()
		WHERE token_hash = $1 AND accepted_at IS NULL AND expires_at > now()
		RETURNING id, practice_id, email, role, invited_by_user_id, accepted_at, expires_at, created_at
	`, tokenHash).Scan(&invite.ID, &invite.PracticeID, &invite.Email, &invite.Role, &invite.InvitedBy, &invite.AcceptedAt, &invite.ExpiresAt, &invite.CreatedAt)
	if err != nil {
		return domain.PracticeInvite{}, false
	}
	_, err = tx.ExecContext(context.Background(), `
		INSERT INTO practice_members (practice_id, user_id, role, active)
		VALUES ($1, $2, $3, true)
		ON CONFLICT (practice_id, user_id)
		DO UPDATE SET role = EXCLUDED.role, active = true, updated_at = now()
	`, invite.PracticeID, userID, invite.Role)
	if err != nil {
		panic(err)
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}
	return invite, true
}

func (s *PostgresStore) CreateUser(email, displayName, passwordHash string) domain.User {
	user := domain.User{ID: id.New(), Email: strings.ToLower(strings.TrimSpace(email)), DisplayName: strings.TrimSpace(displayName), CreatedAt: time.Now().UTC()}
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	defer rollback(tx)
	_, err = tx.ExecContext(context.Background(), `
		INSERT INTO users (id, email, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
	`, user.ID, user.Email, user.DisplayName, user.CreatedAt)
	if err != nil {
		panic(err)
	}
	_, err = tx.ExecContext(context.Background(), `
		INSERT INTO user_credentials (user_id, password_hash, created_at, updated_at)
		VALUES ($1, $2, now(), now())
	`, user.ID, passwordHash)
	if err != nil {
		panic(err)
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}
	return user
}

func (s *PostgresStore) FindUserByEmail(email string) (domain.User, string, bool) {
	var user domain.User
	var passwordHash string
	var emailVerifiedAt sql.NullTime
	err := s.db.QueryRowContext(context.Background(), `
		SELECT u.id, u.email, u.display_name, u.email_verified_at, u.created_at, c.password_hash
		FROM users u
		JOIN user_credentials c ON c.user_id = u.id
		WHERE u.email = $1
	`, strings.ToLower(strings.TrimSpace(email))).Scan(&user.ID, &user.Email, &user.DisplayName, &emailVerifiedAt, &user.CreatedAt, &passwordHash)
	if err != nil {
		return domain.User{}, "", false
	}
	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = &emailVerifiedAt.Time
	}
	return user, passwordHash, true
}

func (s *PostgresStore) CreateEmailVerificationToken(userID, tokenHash string, expiresAt time.Time) {
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO email_verification_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, now())
	`, id.New(), userID, tokenHash, expiresAt)
	if err != nil {
		panic(err)
	}
}

func (s *PostgresStore) VerifyEmailByTokenHash(tokenHash string) (domain.User, bool) {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	defer rollback(tx)
	var userID string
	err = tx.QueryRowContext(context.Background(), `
		UPDATE email_verification_tokens
		SET used_at = now()
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now()
		RETURNING user_id
	`, tokenHash).Scan(&userID)
	if err != nil {
		return domain.User{}, false
	}
	var user domain.User
	var emailVerifiedAt sql.NullTime
	err = tx.QueryRowContext(context.Background(), `
		UPDATE users
		SET email_verified_at = now(), updated_at = now()
		WHERE id = $1
		RETURNING id, email, display_name, email_verified_at, created_at
	`, userID).Scan(&user.ID, &user.Email, &user.DisplayName, &emailVerifiedAt, &user.CreatedAt)
	if err != nil {
		panic(err)
	}
	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = &emailVerifiedAt.Time
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}
	return user, true
}

func (s *PostgresStore) CreatePasswordResetToken(email, tokenHash string, expiresAt time.Time) (domain.User, bool) {
	user, _, ok := s.FindUserByEmail(email)
	if !ok {
		return domain.User{}, false
	}
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, now())
	`, id.New(), user.ID, tokenHash, expiresAt)
	if err != nil {
		panic(err)
	}
	return user, true
}

func (s *PostgresStore) ResetPasswordByTokenHash(tokenHash, passwordHash string) (domain.User, bool) {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	defer rollback(tx)
	var userID string
	err = tx.QueryRowContext(context.Background(), `
		UPDATE password_reset_tokens
		SET used_at = now()
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now()
		RETURNING user_id
	`, tokenHash).Scan(&userID)
	if err != nil {
		return domain.User{}, false
	}
	_, err = tx.ExecContext(context.Background(), `
		UPDATE user_credentials SET password_hash = $2, updated_at = now() WHERE user_id = $1
	`, userID, passwordHash)
	if err != nil {
		panic(err)
	}
	var user domain.User
	var emailVerifiedAt sql.NullTime
	err = tx.QueryRowContext(context.Background(), `
		SELECT id, email, display_name, email_verified_at, created_at
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.DisplayName, &emailVerifiedAt, &user.CreatedAt)
	if err != nil {
		panic(err)
	}
	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = &emailVerifiedAt.Time
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}
	return user, true
}

func (s *PostgresStore) CreateSession(userID, tokenHash string, expiresAt time.Time) Session {
	session := Session{ID: id.New(), UserID: userID, TokenHash: tokenHash, ExpiresAt: expiresAt}
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO user_sessions (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, now())
	`, session.ID, session.UserID, session.TokenHash, session.ExpiresAt)
	if err != nil {
		panic(err)
	}
	return session
}

func (s *PostgresStore) FindSession(tokenHash string) (Session, bool) {
	var session Session
	err := s.db.QueryRowContext(context.Background(), `
		SELECT id, user_id, token_hash, expires_at
		FROM user_sessions
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()
	`, tokenHash).Scan(&session.ID, &session.UserID, &session.TokenHash, &session.ExpiresAt)
	return session, err == nil
}

func (s *PostgresStore) CreatePractice(name, ownerUserID string) domain.Practice {
	practice := domain.Practice{ID: id.New(), Name: strings.TrimSpace(name), Specialty: "dental", CreatedAt: time.Now().UTC()}
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	defer rollback(tx)
	_, err = tx.ExecContext(context.Background(), `
		INSERT INTO practices (id, name, specialty, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
	`, practice.ID, practice.Name, practice.Specialty, practice.CreatedAt)
	if err != nil {
		panic(err)
	}
	_, err = tx.ExecContext(context.Background(), `
		INSERT INTO practice_members (practice_id, user_id, role, active)
		VALUES ($1, $2, 'owner', true)
	`, practice.ID, ownerUserID)
	if err != nil {
		panic(err)
	}
	for _, role := range defaultRoles(practice.ID) {
		if err := insertRoleTx(tx, role); err != nil {
			panic(err)
		}
	}
	config := domain.AssistantConfig{
		ID:                id.New(),
		PracticeID:        practice.ID,
		Greeting:          "Thank you for calling " + practice.Name + ". How can I help you today?",
		EscalationPhone:   "",
		NotificationEmail: "",
		Settings:          map[string]string{"specialty": "dental"},
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := insertAssistantConfigTx(tx, config); err != nil {
		panic(err)
	}
	voiceConfig := domain.VoiceProviderConfig{
		ID:            id.New(),
		PracticeID:    practice.ID,
		Provider:      "vapi",
		WebhookStatus: "not_configured",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	if err := insertVoiceProviderConfigTx(tx, voiceConfig); err != nil {
		panic(err)
	}
	calendarConfig := domain.CalendarConfig{
		ID:         id.New(),
		PracticeID: practice.ID,
		Mode:       "request_only",
		Provider:   "none",
		Timezone:   "America/New_York",
		Status:     "not_configured",
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := insertCalendarConfigTx(tx, calendarConfig); err != nil {
		panic(err)
	}
	if err := insertBillingSubscriptionTx(tx, defaultBillingSubscription(practice.ID)); err != nil {
		panic(err)
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}
	return practice
}

func (s *PostgresStore) ListPracticesForUser(userID string) []domain.Practice {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT p.id, p.name, p.specialty, p.created_at
		FROM practices p
		JOIN practice_members m ON m.practice_id = p.id
		WHERE m.user_id = $1 AND m.active = true
		ORDER BY p.created_at DESC
	`, userID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	out := make([]domain.Practice, 0)
	for rows.Next() {
		var practice domain.Practice
		if err := rows.Scan(&practice.ID, &practice.Name, &practice.Specialty, &practice.CreatedAt); err != nil {
			panic(err)
		}
		out = append(out, practice)
	}
	return out
}

func (s *PostgresStore) IsMember(practiceID, userID string) bool {
	var exists bool
	err := s.db.QueryRowContext(context.Background(), `
		SELECT EXISTS(SELECT 1 FROM practice_members WHERE practice_id = $1 AND user_id = $2 AND active = true)
	`, practiceID, userID).Scan(&exists)
	return err == nil && exists
}

func (s *PostgresStore) GetMember(practiceID, userID string) (domain.PracticeMember, bool) {
	var member domain.PracticeMember
	err := s.db.QueryRowContext(context.Background(), `
		SELECT m.practice_id, m.user_id, m.role, u.email, u.display_name, m.active
		FROM practice_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.practice_id = $1 AND m.user_id = $2 AND m.active = true
	`, practiceID, userID).Scan(&member.PracticeID, &member.UserID, &member.Role, &member.Email, &member.DisplayName, &member.Active)
	return member, err == nil
}

func (s *PostgresStore) ListMembers(practiceID string) []domain.PracticeMember {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT m.practice_id, m.user_id, m.role, u.email, u.display_name, m.active
		FROM practice_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.practice_id = $1 AND m.active = true
		ORDER BY u.display_name, u.email
	`, practiceID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	out := make([]domain.PracticeMember, 0)
	for rows.Next() {
		var member domain.PracticeMember
		if err := rows.Scan(&member.PracticeID, &member.UserID, &member.Role, &member.Email, &member.DisplayName, &member.Active); err != nil {
			panic(err)
		}
		out = append(out, member)
	}
	return out
}

func (s *PostgresStore) UpsertMember(practiceID, userID, role string) domain.PracticeMember {
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO practice_members (practice_id, user_id, role, active)
		VALUES ($1, $2, $3, true)
		ON CONFLICT (practice_id, user_id)
		DO UPDATE SET role = EXCLUDED.role, active = true, updated_at = now()
	`, practiceID, userID, role)
	if err != nil {
		panic(err)
	}
	member, _ := s.GetMember(practiceID, userID)
	return member
}

func (s *PostgresStore) UpdateMemberRole(practiceID, userID, role string) (domain.PracticeMember, bool) {
	result, err := s.db.ExecContext(context.Background(), `
		UPDATE practice_members SET role = $3, updated_at = now()
		WHERE practice_id = $1 AND user_id = $2 AND active = true
	`, practiceID, userID, role)
	if err != nil {
		panic(err)
	}
	if affected(result) == 0 {
		return domain.PracticeMember{}, false
	}
	return s.GetMember(practiceID, userID)
}

func (s *PostgresStore) DisableMember(practiceID, userID string) bool {
	result, err := s.db.ExecContext(context.Background(), `
		UPDATE practice_members SET active = false, updated_at = now()
		WHERE practice_id = $1 AND user_id = $2
	`, practiceID, userID)
	if err != nil {
		panic(err)
	}
	return affected(result) > 0
}

func (s *PostgresStore) ListRoles(practiceID string) []domain.Role {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id, practice_id, name, description, system, created_at
		FROM roles
		WHERE practice_id = $1
		ORDER BY system DESC, name
	`, practiceID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	roles := make([]domain.Role, 0)
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.PracticeID, &role.Name, &role.Description, &role.System, &role.CreatedAt); err != nil {
			panic(err)
		}
		role.Permissions = s.permissionsForRole(role.ID)
		roles = append(roles, role)
	}
	return roles
}

func (s *PostgresStore) GetRole(practiceID, name string) (domain.Role, bool) {
	var role domain.Role
	err := s.db.QueryRowContext(context.Background(), `
		SELECT id, practice_id, name, description, system, created_at
		FROM roles
		WHERE practice_id = $1 AND name = $2
	`, practiceID, strings.ToLower(strings.TrimSpace(name))).Scan(&role.ID, &role.PracticeID, &role.Name, &role.Description, &role.System, &role.CreatedAt)
	if err != nil {
		return domain.Role{}, false
	}
	role.Permissions = s.permissionsForRole(role.ID)
	return role, true
}

func (s *PostgresStore) CreateRole(practiceID, name, description string, permissions []string) domain.Role {
	role := domain.Role{
		ID:          id.New(),
		PracticeID:  practiceID,
		Name:        strings.ToLower(strings.TrimSpace(name)),
		Description: strings.TrimSpace(description),
		System:      false,
		Permissions: permissions,
		CreatedAt:   time.Now().UTC(),
	}
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	defer rollback(tx)
	if err := insertRoleTx(tx, role); err != nil {
		panic(err)
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}
	return role
}

func (s *PostgresStore) CreateLocation(location domain.Location) domain.Location {
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
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO locations (id, practice_id, name, address_line1, address_line2, city, region, postal_code, country, timezone, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, location.ID, location.PracticeID, location.Name, location.AddressLine1, location.AddressLine2, location.City, location.Region, location.PostalCode, location.Country, location.Timezone, location.CreatedAt)
	if err != nil {
		panic(err)
	}
	return location
}

func (s *PostgresStore) ListLocations(practiceID string) []domain.Location {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id, practice_id, name, address_line1, address_line2, city, region, postal_code, country, timezone, created_at
		FROM locations
		WHERE practice_id = $1
		ORDER BY created_at DESC
	`, practiceID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	out := make([]domain.Location, 0)
	for rows.Next() {
		var location domain.Location
		if err := rows.Scan(&location.ID, &location.PracticeID, &location.Name, &location.AddressLine1, &location.AddressLine2, &location.City, &location.Region, &location.PostalCode, &location.Country, &location.Timezone, &location.CreatedAt); err != nil {
			panic(err)
		}
		out = append(out, location)
	}
	return out
}

func (s *PostgresStore) HasPermission(practiceID, userID, permission string) bool {
	var exists bool
	err := s.db.QueryRowContext(context.Background(), `
		SELECT EXISTS(
			SELECT 1
			FROM practice_members m
			JOIN roles r ON r.practice_id = m.practice_id AND r.name = m.role
			JOIN role_permissions rp ON rp.role_id = r.id
			WHERE m.practice_id = $1 AND m.user_id = $2 AND m.active = true AND rp.permission = $3
		)
	`, practiceID, userID, permission).Scan(&exists)
	return err == nil && exists
}

func (s *PostgresStore) GetAssistantConfig(practiceID string) (domain.AssistantConfig, bool) {
	var config domain.AssistantConfig
	var settings []byte
	err := s.db.QueryRowContext(context.Background(), `
		SELECT id, practice_id, greeting, escalation_phone, notification_email, config_json, created_at, updated_at
		FROM assistant_configs
		WHERE practice_id = $1
	`, practiceID).Scan(&config.ID, &config.PracticeID, &config.Greeting, &config.EscalationPhone, &config.NotificationEmail, &settings, &config.CreatedAt, &config.UpdatedAt)
	if err != nil {
		return domain.AssistantConfig{}, false
	}
	_ = json.Unmarshal(settings, &config.Settings)
	return config, true
}

func (s *PostgresStore) SaveAssistantConfig(config domain.AssistantConfig) domain.AssistantConfig {
	settings, _ := json.Marshal(config.Settings)
	config.UpdatedAt = time.Now().UTC()
	_, err := s.db.ExecContext(context.Background(), `
		UPDATE assistant_configs
		SET greeting = $2, escalation_phone = $3, notification_email = $4, config_json = $5, updated_at = $6
		WHERE practice_id = $1
	`, config.PracticeID, config.Greeting, config.EscalationPhone, config.NotificationEmail, settings, config.UpdatedAt)
	if err != nil {
		panic(err)
	}
	return config
}

func (s *PostgresStore) GetVoiceProviderConfig(practiceID string) (domain.VoiceProviderConfig, bool) {
	var config domain.VoiceProviderConfig
	var lastWebhookAt sql.NullTime
	err := s.db.QueryRowContext(context.Background(), `
		SELECT id, practice_id, provider, phone_number, assistant_id, webhook_status, last_webhook_at, created_at, updated_at
		FROM voice_provider_configs
		WHERE practice_id = $1
	`, practiceID).Scan(&config.ID, &config.PracticeID, &config.Provider, &config.PhoneNumber, &config.AssistantID, &config.WebhookStatus, &lastWebhookAt, &config.CreatedAt, &config.UpdatedAt)
	if err != nil {
		return domain.VoiceProviderConfig{}, false
	}
	if lastWebhookAt.Valid {
		config.LastWebhookAt = &lastWebhookAt.Time
	}
	return config, true
}

func (s *PostgresStore) SaveVoiceProviderConfig(config domain.VoiceProviderConfig) domain.VoiceProviderConfig {
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
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO voice_provider_configs (id, practice_id, provider, phone_number, assistant_id, webhook_status, last_webhook_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (practice_id)
		DO UPDATE SET provider = EXCLUDED.provider,
			phone_number = EXCLUDED.phone_number,
			assistant_id = EXCLUDED.assistant_id,
			webhook_status = EXCLUDED.webhook_status,
			last_webhook_at = EXCLUDED.last_webhook_at,
			updated_at = EXCLUDED.updated_at
	`, config.ID, config.PracticeID, config.Provider, config.PhoneNumber, config.AssistantID, config.WebhookStatus, config.LastWebhookAt, config.CreatedAt, config.UpdatedAt)
	if err != nil {
		panic(err)
	}
	saved, _ := s.GetVoiceProviderConfig(config.PracticeID)
	return saved
}

func (s *PostgresStore) GetCalendarConfig(practiceID string) (domain.CalendarConfig, bool) {
	var config domain.CalendarConfig
	var tokenExpiresAt sql.NullTime
	err := s.db.QueryRowContext(context.Background(), `
		SELECT id, practice_id, mode, provider, booking_url, calendar_id, timezone, status, instructions, oauth_connected, oauth_access_token_enc, oauth_refresh_token_enc, oauth_token_expires_at, created_at, updated_at
		FROM calendar_configs
		WHERE practice_id = $1
	`, practiceID).Scan(&config.ID, &config.PracticeID, &config.Mode, &config.Provider, &config.BookingURL, &config.CalendarID, &config.Timezone, &config.Status, &config.Instructions, &config.OAuthConnected, &config.OAuthAccessTokenEnc, &config.OAuthRefreshTokenEnc, &tokenExpiresAt, &config.CreatedAt, &config.UpdatedAt)
	if err != nil {
		return domain.CalendarConfig{}, false
	}
	if tokenExpiresAt.Valid {
		config.OAuthTokenExpiresAt = &tokenExpiresAt.Time
	}
	return config, true
}

func (s *PostgresStore) SaveCalendarConfig(config domain.CalendarConfig) domain.CalendarConfig {
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
	now := time.Now().UTC()
	if config.CreatedAt.IsZero() {
		config.CreatedAt = now
	}
	config.UpdatedAt = now
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO calendar_configs (id, practice_id, mode, provider, booking_url, calendar_id, timezone, status, instructions, oauth_connected, oauth_access_token_enc, oauth_refresh_token_enc, oauth_token_expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, false, '', '', NULL, $10, $11)
		ON CONFLICT (practice_id)
		DO UPDATE SET mode = EXCLUDED.mode,
			provider = EXCLUDED.provider,
			booking_url = EXCLUDED.booking_url,
			calendar_id = EXCLUDED.calendar_id,
			timezone = EXCLUDED.timezone,
			status = CASE WHEN calendar_configs.oauth_connected THEN 'connected' ELSE EXCLUDED.status END,
			instructions = EXCLUDED.instructions,
			updated_at = EXCLUDED.updated_at
	`, config.ID, config.PracticeID, config.Mode, config.Provider, config.BookingURL, config.CalendarID, config.Timezone, config.Status, config.Instructions, config.CreatedAt, config.UpdatedAt)
	if err != nil {
		panic(err)
	}
	saved, _ := s.GetCalendarConfig(config.PracticeID)
	return saved
}

func (s *PostgresStore) SaveCalendarOAuth(practiceID, accessTokenEnc, refreshTokenEnc string, expiresAt time.Time) (domain.CalendarConfig, bool) {
	_, err := s.db.ExecContext(context.Background(), `
		UPDATE calendar_configs
		SET mode = 'google',
			provider = 'google',
			status = 'connected',
			oauth_connected = true,
			oauth_access_token_enc = $2,
			oauth_refresh_token_enc = CASE WHEN $3 = '' THEN oauth_refresh_token_enc ELSE $3 END,
			oauth_token_expires_at = $4,
			updated_at = $5
		WHERE practice_id = $1
	`, practiceID, accessTokenEnc, refreshTokenEnc, expiresAt, time.Now().UTC())
	if err != nil {
		panic(err)
	}
	return s.GetCalendarConfig(practiceID)
}

func (s *PostgresStore) GetBillingSubscription(practiceID string) (domain.BillingSubscription, bool) {
	var subscription domain.BillingSubscription
	var trialEndsAt sql.NullTime
	var currentPeriodEndsAt sql.NullTime
	err := s.db.QueryRowContext(context.Background(), `
		SELECT id, practice_id, plan, status, included_minutes, overage_cents, stripe_customer_id, stripe_subscription_id, trial_ends_at, current_period_ends_at, created_at, updated_at
		FROM billing_subscriptions
		WHERE practice_id = $1
	`, practiceID).Scan(&subscription.ID, &subscription.PracticeID, &subscription.Plan, &subscription.Status, &subscription.IncludedMinutes, &subscription.OverageCents, &subscription.StripeCustomerID, &subscription.StripeSubscriptionID, &trialEndsAt, &currentPeriodEndsAt, &subscription.CreatedAt, &subscription.UpdatedAt)
	if err != nil {
		return domain.BillingSubscription{}, false
	}
	if trialEndsAt.Valid {
		subscription.TrialEndsAt = &trialEndsAt.Time
	}
	if currentPeriodEndsAt.Valid {
		subscription.CurrentPeriodEndsAt = &currentPeriodEndsAt.Time
	}
	return subscription, true
}

func (s *PostgresStore) SaveBillingSubscription(subscription domain.BillingSubscription) domain.BillingSubscription {
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
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO billing_subscriptions (id, practice_id, plan, status, included_minutes, overage_cents, stripe_customer_id, stripe_subscription_id, trial_ends_at, current_period_ends_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (practice_id)
		DO UPDATE SET plan = EXCLUDED.plan,
			status = EXCLUDED.status,
			included_minutes = EXCLUDED.included_minutes,
			overage_cents = EXCLUDED.overage_cents,
			stripe_customer_id = EXCLUDED.stripe_customer_id,
			stripe_subscription_id = EXCLUDED.stripe_subscription_id,
			trial_ends_at = EXCLUDED.trial_ends_at,
			current_period_ends_at = EXCLUDED.current_period_ends_at,
			updated_at = EXCLUDED.updated_at
	`, subscription.ID, subscription.PracticeID, subscription.Plan, subscription.Status, subscription.IncludedMinutes, subscription.OverageCents, subscription.StripeCustomerID, subscription.StripeSubscriptionID, subscription.TrialEndsAt, subscription.CurrentPeriodEndsAt, subscription.CreatedAt, subscription.UpdatedAt)
	if err != nil {
		panic(err)
	}
	saved, _ := s.GetBillingSubscription(subscription.PracticeID)
	return saved
}

func (s *PostgresStore) CreateCall(call domain.CallSession) domain.CallSession {
	if call.ID == "" {
		call.ID = id.New()
	}
	if call.StartedAt.IsZero() {
		call.StartedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO call_sessions (id, practice_id, provider, provider_call_id, caller_phone, status, started_at, ended_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, call.ID, call.PracticeID, call.Provider, call.ProviderCallID, call.CallerPhone, call.Status, call.StartedAt, call.EndedAt)
	if err != nil {
		panic(err)
	}
	return call
}

func (s *PostgresStore) CreateSummary(summary domain.CallSummary) domain.CallSummary {
	if summary.ID == "" {
		summary.ID = id.New()
	}
	if summary.CreatedAt.IsZero() {
		summary.CreatedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO call_summaries (id, call_session_id, practice_id, caller_name, reason, urgency, ai_action, follow_up_needed, summary, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, summary.ID, nullString(summary.CallSessionID), summary.PracticeID, summary.CallerName, summary.Reason, summary.Urgency, summary.AIAction, summary.FollowUpNeeded, summary.Summary, summary.CreatedAt)
	if err != nil {
		panic(err)
	}
	return summary
}

func (s *PostgresStore) ListSummaries(practiceID string) []domain.CallSummary {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id, call_session_id, practice_id, caller_name, reason, urgency, ai_action, follow_up_needed, summary, created_at
		FROM call_summaries
		WHERE practice_id = $1
		ORDER BY created_at DESC
	`, practiceID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	out := make([]domain.CallSummary, 0)
	for rows.Next() {
		var summary domain.CallSummary
		var callSessionID sql.NullString
		if err := rows.Scan(&summary.ID, &callSessionID, &summary.PracticeID, &summary.CallerName, &summary.Reason, &summary.Urgency, &summary.AIAction, &summary.FollowUpNeeded, &summary.Summary, &summary.CreatedAt); err != nil {
			panic(err)
		}
		summary.CallSessionID = callSessionID.String
		out = append(out, summary)
	}
	return out
}

func (s *PostgresStore) CreateAppointment(request domain.AppointmentRequest) domain.AppointmentRequest {
	if request.ID == "" {
		request.ID = id.New()
	}
	if request.CreatedAt.IsZero() {
		request.CreatedAt = time.Now().UTC()
	}
	if request.Status == "" {
		request.Status = "new"
	}
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO appointment_requests (id, practice_id, call_session_id, caller_name, caller_phone, request_type, preferred_time, insurance, notes, staff_note, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, request.ID, request.PracticeID, nullString(request.CallSessionID), request.CallerName, request.CallerPhone, request.RequestType, request.PreferredTime, request.Insurance, request.Notes, request.StaffNote, request.Status, request.CreatedAt)
	if err != nil {
		panic(err)
	}
	return request
}

func (s *PostgresStore) ListAppointments(practiceID string) []domain.AppointmentRequest {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id, practice_id, call_session_id, caller_name, caller_phone, request_type, preferred_time, insurance, notes, staff_note, status, created_at
		FROM appointment_requests
		WHERE practice_id = $1
		ORDER BY created_at DESC
	`, practiceID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	out := make([]domain.AppointmentRequest, 0)
	for rows.Next() {
		var request domain.AppointmentRequest
		var callSessionID sql.NullString
		if err := rows.Scan(&request.ID, &request.PracticeID, &callSessionID, &request.CallerName, &request.CallerPhone, &request.RequestType, &request.PreferredTime, &request.Insurance, &request.Notes, &request.StaffNote, &request.Status, &request.CreatedAt); err != nil {
			panic(err)
		}
		request.CallSessionID = callSessionID.String
		out = append(out, request)
	}
	return out
}

func (s *PostgresStore) UpdateAppointmentRequest(practiceID, requestID, status, staffNote string) (domain.AppointmentRequest, bool) {
	var request domain.AppointmentRequest
	var callSessionID sql.NullString
	err := s.db.QueryRowContext(context.Background(), `
		UPDATE appointment_requests
		SET status = CASE WHEN $3 = '' THEN status ELSE $3 END,
		    staff_note = $4
		WHERE practice_id = $1 AND id = $2
		RETURNING id, practice_id, call_session_id, caller_name, caller_phone, request_type, preferred_time, insurance, notes, staff_note, status, created_at
	`, practiceID, requestID, status, staffNote).Scan(&request.ID, &request.PracticeID, &callSessionID, &request.CallerName, &request.CallerPhone, &request.RequestType, &request.PreferredTime, &request.Insurance, &request.Notes, &request.StaffNote, &request.Status, &request.CreatedAt)
	if err == sql.ErrNoRows {
		return domain.AppointmentRequest{}, false
	}
	if err != nil {
		panic(err)
	}
	request.CallSessionID = callSessionID.String
	return request, true
}

func (s *PostgresStore) CreateAuditLog(log domain.AuditLog) domain.AuditLog {
	if log.ID == "" {
		log.ID = id.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}
	if log.Metadata == nil {
		log.Metadata = map[string]string{}
	}
	metadata, _ := json.Marshal(log.Metadata)
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO audit_logs (id, practice_id, actor_type, actor_id, action, target_type, target_id, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, log.ID, nullString(log.PracticeID), log.ActorType, log.ActorID, log.Action, log.TargetType, log.TargetID, metadata, log.CreatedAt)
	if err != nil {
		panic(err)
	}
	return log
}

func (s *PostgresStore) ListAuditLogs(practiceID string, limit int) []domain.AuditLog {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id, COALESCE(practice_id::text, ''), actor_type, actor_id, action, target_type, target_id, metadata, created_at
		FROM audit_logs
		WHERE practice_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, practiceID, limit)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	out := make([]domain.AuditLog, 0)
	for rows.Next() {
		var log domain.AuditLog
		var metadata []byte
		if err := rows.Scan(&log.ID, &log.PracticeID, &log.ActorType, &log.ActorID, &log.Action, &log.TargetType, &log.TargetID, &metadata, &log.CreatedAt); err != nil {
			panic(err)
		}
		_ = json.Unmarshal(metadata, &log.Metadata)
		out = append(out, log)
	}
	return out
}

func insertRoleTx(tx *sql.Tx, role domain.Role) error {
	_, err := tx.ExecContext(context.Background(), `
		INSERT INTO roles (id, practice_id, name, description, system, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, role.ID, role.PracticeID, role.Name, role.Description, role.System, role.CreatedAt)
	if err != nil {
		return err
	}
	for _, permission := range role.Permissions {
		if _, err := tx.ExecContext(context.Background(), `
			INSERT INTO role_permissions (role_id, permission)
			VALUES ($1, $2)
		`, role.ID, permission); err != nil {
			return err
		}
	}
	return nil
}

func insertAssistantConfigTx(tx *sql.Tx, config domain.AssistantConfig) error {
	settings, _ := json.Marshal(config.Settings)
	_, err := tx.ExecContext(context.Background(), `
		INSERT INTO assistant_configs (id, practice_id, greeting, escalation_phone, notification_email, config_json, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, config.ID, config.PracticeID, config.Greeting, config.EscalationPhone, config.NotificationEmail, settings, config.CreatedAt, config.UpdatedAt)
	return err
}

func insertVoiceProviderConfigTx(tx *sql.Tx, config domain.VoiceProviderConfig) error {
	_, err := tx.ExecContext(context.Background(), `
		INSERT INTO voice_provider_configs (id, practice_id, provider, phone_number, assistant_id, webhook_status, last_webhook_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, config.ID, config.PracticeID, config.Provider, config.PhoneNumber, config.AssistantID, config.WebhookStatus, config.LastWebhookAt, config.CreatedAt, config.UpdatedAt)
	return err
}

func insertCalendarConfigTx(tx *sql.Tx, config domain.CalendarConfig) error {
	_, err := tx.ExecContext(context.Background(), `
		INSERT INTO calendar_configs (id, practice_id, mode, provider, booking_url, calendar_id, timezone, status, instructions, oauth_connected, oauth_access_token_enc, oauth_refresh_token_enc, oauth_token_expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, false, '', '', NULL, $10, $11)
	`, config.ID, config.PracticeID, config.Mode, config.Provider, config.BookingURL, config.CalendarID, config.Timezone, config.Status, config.Instructions, config.CreatedAt, config.UpdatedAt)
	return err
}

func insertBillingSubscriptionTx(tx *sql.Tx, subscription domain.BillingSubscription) error {
	_, err := tx.ExecContext(context.Background(), `
		INSERT INTO billing_subscriptions (id, practice_id, plan, status, included_minutes, overage_cents, stripe_customer_id, stripe_subscription_id, trial_ends_at, current_period_ends_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, subscription.ID, subscription.PracticeID, subscription.Plan, subscription.Status, subscription.IncludedMinutes, subscription.OverageCents, subscription.StripeCustomerID, subscription.StripeSubscriptionID, subscription.TrialEndsAt, subscription.CurrentPeriodEndsAt, subscription.CreatedAt, subscription.UpdatedAt)
	return err
}

func (s *PostgresStore) permissionsForRole(roleID string) []string {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT permission FROM role_permissions WHERE role_id = $1 ORDER BY permission
	`, roleID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			panic(err)
		}
		out = append(out, permission)
	}
	return out
}

func scanInvites(rows *sql.Rows) []domain.PracticeInvite {
	out := make([]domain.PracticeInvite, 0)
	for rows.Next() {
		var invite domain.PracticeInvite
		var acceptedAt sql.NullTime
		if err := rows.Scan(&invite.ID, &invite.PracticeID, &invite.Email, &invite.Role, &invite.InvitedBy, &acceptedAt, &invite.ExpiresAt, &invite.CreatedAt); err != nil {
			panic(err)
		}
		if acceptedAt.Valid {
			invite.AcceptedAt = &acceptedAt.Time
		}
		out = append(out, invite)
	}
	return out
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func affected(result sql.Result) int64 {
	count, err := result.RowsAffected()
	if err != nil {
		return 0
	}
	return count
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}
