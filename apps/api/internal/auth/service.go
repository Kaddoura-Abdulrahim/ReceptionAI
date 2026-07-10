package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode"

	"dentaldesk/apps/api/internal/domain"
	"dentaldesk/apps/api/internal/store"
)

const SessionCookieName = "dd_session"
const CSRFCookieName = "dd_csrf"

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrEmailTaken = errors.New("email already registered")
var ErrWeakPassword = errors.New("password must be at least 12 characters and include uppercase, lowercase, and a number")

type Service struct {
	store         store.Store
	sessionSecret string
	secureCookies bool
}

type Actor struct {
	Type       string   `json:"type"`
	ID         string   `json:"id"`
	PracticeID string   `json:"practiceId,omitempty"`
	Role       string   `json:"role,omitempty"`
	Scopes     []string `json:"scopes,omitempty"`
}

func NewService(store store.Store, sessionSecret string, secureCookies bool) *Service {
	return &Service{store: store, sessionSecret: sessionSecret, secureCookies: secureCookies}
}

func (s *Service) Register(ctx context.Context, email, displayName, password string) (domain.User, string, error) {
	_ = ctx
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return domain.User{}, "", ErrInvalidCredentials
	}
	if err := ValidatePassword(password); err != nil {
		return domain.User{}, "", err
	}
	if _, _, ok := s.store.FindUserByEmail(email); ok {
		return domain.User{}, "", ErrEmailTaken
	}
	hash, err := hashPassword(password)
	if err != nil {
		return domain.User{}, "", err
	}
	user := s.store.CreateUser(email, displayName, hash)
	token, err := randomToken(32)
	if err != nil {
		return domain.User{}, "", err
	}
	s.store.CreateSession(user.ID, s.hashToken(token), time.Now().UTC().Add(30*24*time.Hour))
	return user, token, nil
}

func (s *Service) CreateUser(ctx context.Context, email, displayName, password string) (domain.User, error) {
	_ = ctx
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return domain.User{}, ErrInvalidCredentials
	}
	if err := ValidatePassword(password); err != nil {
		return domain.User{}, err
	}
	if _, _, ok := s.store.FindUserByEmail(email); ok {
		return domain.User{}, ErrEmailTaken
	}
	hash, err := hashPassword(password)
	if err != nil {
		return domain.User{}, err
	}
	return s.store.CreateUser(email, displayName, hash), nil
}

func (s *Service) Login(ctx context.Context, email, password string) (domain.User, string, error) {
	_ = ctx
	user, passwordHash, ok := s.store.FindUserByEmail(email)
	if !ok || !verifyPassword(password, passwordHash) {
		return domain.User{}, "", ErrInvalidCredentials
	}
	token, err := randomToken(32)
	if err != nil {
		return domain.User{}, "", err
	}
	s.store.CreateSession(user.ID, s.hashToken(token), time.Now().UTC().Add(30*24*time.Hour))
	return user, token, nil
}

func (s *Service) CreateEmailVerification(ctx context.Context, userID string) (string, error) {
	_ = ctx
	token, err := randomToken(32)
	if err != nil {
		return "", err
	}
	s.store.CreateEmailVerificationToken(userID, s.hashToken(token), time.Now().UTC().Add(24*time.Hour))
	return token, nil
}

func (s *Service) VerifyEmail(ctx context.Context, token string) (domain.User, bool) {
	_ = ctx
	return s.store.VerifyEmailByTokenHash(s.hashToken(token))
}

func (s *Service) CreatePasswordReset(ctx context.Context, email string) (domain.User, string, bool, error) {
	_ = ctx
	token, err := randomToken(32)
	if err != nil {
		return domain.User{}, "", false, err
	}
	user, ok := s.store.CreatePasswordResetToken(email, s.hashToken(token), time.Now().UTC().Add(time.Hour))
	return user, token, ok, nil
}

func (s *Service) ResetPassword(ctx context.Context, token, password string) (domain.User, bool, error) {
	_ = ctx
	if err := ValidatePassword(password); err != nil {
		return domain.User{}, false, err
	}
	hash, err := hashPassword(password)
	if err != nil {
		return domain.User{}, false, err
	}
	user, ok := s.store.ResetPasswordByTokenHash(s.hashToken(token), hash)
	return user, ok, nil
}

func (s *Service) UserIDFromRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return "", false
	}
	session, ok := s.store.FindSession(s.hashToken(cookie.Value))
	if !ok {
		return "", false
	}
	return session.UserID, true
}

func (s *Service) SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: s.sameSiteMode(),
		Secure:   s.secureCookies,
		Expires:  time.Now().UTC().Add(30 * 24 * time.Hour),
	})
}

func (s *Service) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: s.sameSiteMode(),
		Secure:   s.secureCookies,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func (s *Service) sameSiteMode() http.SameSite {
	if s.secureCookies {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func (s *Service) NewCSRFToken() (string, error) {
	return randomToken(32)
}

func (s *Service) hashToken(token string) string {
	mac := hmac.New(sha256.New, []byte(s.sessionSecret))
	mac.Write([]byte(token))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Service) NewPublicToken() (string, error) {
	return randomToken(32)
}

func (s *Service) HashPublicToken(token string) string {
	return s.hashToken(token)
}

func hashPassword(password string) (string, error) {
	salt, err := randomBytes(16)
	if err != nil {
		return "", err
	}
	key := pbkdf2SHA256([]byte(password), salt, 120000, 32)
	return base64.RawURLEncoding.EncodeToString(salt) + "." + base64.RawURLEncoding.EncodeToString(key), nil
}

func verifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, ".")
	if len(parts) != 2 {
		return false
	}
	salt, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	want, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	got := pbkdf2SHA256([]byte(password), salt, 120000, len(want))
	return hmac.Equal(got, want)
}

func pbkdf2SHA256(password, salt []byte, iter, keyLen int) []byte {
	hashLen := 32
	numBlocks := (keyLen + hashLen - 1) / hashLen
	out := make([]byte, 0, numBlocks*hashLen)
	for block := 1; block <= numBlocks; block++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		mac.Write([]byte{byte(block >> 24), byte(block >> 16), byte(block >> 8), byte(block)})
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)
		for i := 1; i < iter; i++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t...)
	}
	return out[:keyLen]
}

func randomToken(n int) (string, error) {
	b, err := randomBytes(n)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

func ValidatePassword(password string) error {
	if len(password) < 12 {
		return ErrWeakPassword
	}
	var hasUpper, hasLower, hasNumber bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsNumber(r):
			hasNumber = true
		}
	}
	if !hasUpper || !hasLower || !hasNumber {
		return ErrWeakPassword
	}
	lower := strings.ToLower(password)
	if strings.Contains(lower, "password") || strings.Contains(lower, "change-me") || strings.Contains(lower, "dentaldesk") {
		return ErrWeakPassword
	}
	return nil
}
