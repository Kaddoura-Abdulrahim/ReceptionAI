package app

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"dentaldesk/apps/api/internal/domain"
	"dentaldesk/apps/api/internal/platform/httpx"
)

const googleCalendarScope = "https://www.googleapis.com/auth/calendar.events"

type googleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

func (s *Server) startGoogleCalendarOAuth(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	if s.cfg.GoogleClientID == "" || s.cfg.GoogleClientSecret == "" {
		httpx.Error(w, http.StatusBadRequest, "google oauth credentials are not configured")
		return
	}
	state, err := s.signCalendarState(practiceID, userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not create oauth state")
		return
	}
	values := url.Values{}
	values.Set("client_id", s.cfg.GoogleClientID)
	values.Set("redirect_uri", s.cfg.GoogleRedirectURL)
	values.Set("response_type", "code")
	values.Set("scope", googleCalendarScope)
	values.Set("access_type", "offline")
	values.Set("prompt", "consent")
	values.Set("state", state)
	httpx.JSON(w, http.StatusOK, map[string]string{
		"authorizationUrl": "https://accounts.google.com/o/oauth2/v2/auth?" + values.Encode(),
	})
}

func (s *Server) googleCalendarOAuthCallback(w http.ResponseWriter, r *http.Request, userID string) {
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if code == "" || state == "" {
		httpx.Error(w, http.StatusBadRequest, "missing oauth code or state")
		return
	}
	practiceID, stateUserID, ok := s.verifyCalendarState(state)
	if !ok || stateUserID != userID || !s.store.HasPermission(practiceID, userID, "calendar:update") {
		httpx.Error(w, http.StatusForbidden, "invalid oauth state")
		return
	}
	token, err := s.exchangeGoogleCode(r, code)
	if err != nil {
		httpx.Error(w, http.StatusBadGateway, err.Error())
		return
	}
	if token.AccessToken == "" {
		httpx.Error(w, http.StatusBadGateway, "google did not return an access token")
		return
	}
	accessEnc, err := s.encryptCalendarToken(token.AccessToken)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not protect access token")
		return
	}
	refreshEnc := ""
	if token.RefreshToken != "" {
		refreshEnc, err = s.encryptCalendarToken(token.RefreshToken)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not protect refresh token")
			return
		}
	}
	expiresAt := time.Now().UTC().Add(time.Duration(token.ExpiresIn) * time.Second)
	config, ok := s.store.SaveCalendarOAuth(practiceID, accessEnc, refreshEnc, expiresAt)
	if !ok {
		httpx.Error(w, http.StatusNotFound, "calendar config not found")
		return
	}
	s.audit(practiceID, "user", userID, "calendar.oauth_connected", "calendar_config", config.ID, map[string]string{"provider": "google"})
	http.Redirect(w, r, s.cfg.WebBaseURL+"/?calendar=connected", http.StatusFound)
}

func (s *Server) createGoogleCalendarEvent(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	var req struct {
		Summary     string `json:"summary"`
		Description string `json:"description"`
		Start       string `json:"start"`
		End         string `json:"end"`
		Attendee    string `json:"attendee"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid calendar event")
		return
	}
	if strings.TrimSpace(req.Summary) == "" || strings.TrimSpace(req.Start) == "" || strings.TrimSpace(req.End) == "" {
		httpx.Error(w, http.StatusBadRequest, "summary, start, and end are required")
		return
	}
	start, err := time.Parse(time.RFC3339, req.Start)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "start must be RFC3339")
		return
	}
	end, err := time.Parse(time.RFC3339, req.End)
	if err != nil || !end.After(start) {
		httpx.Error(w, http.StatusBadRequest, "end must be RFC3339 after start")
		return
	}
	config, ok := s.store.GetCalendarConfig(practiceID)
	if !ok || !config.OAuthConnected || config.OAuthAccessTokenEnc == "" {
		httpx.Error(w, http.StatusBadRequest, "google calendar is not connected")
		return
	}
	if strings.TrimSpace(config.CalendarID) == "" {
		httpx.Error(w, http.StatusBadRequest, "calendar ID is required")
		return
	}
	accessToken, err := s.decryptCalendarToken(config.OAuthAccessTokenEnc)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not read calendar token")
		return
	}
	if config.OAuthTokenExpiresAt != nil && time.Now().UTC().After(config.OAuthTokenExpiresAt.Add(-2*time.Minute)) {
		accessToken, config, err = s.refreshGoogleAccessToken(r, config)
		if err != nil {
			httpx.Error(w, http.StatusBadGateway, err.Error())
			return
		}
	}
	event, err := s.postGoogleCalendarEvent(r, config, accessToken, req.Summary, req.Description, req.Start, req.End, req.Attendee)
	if err != nil {
		httpx.Error(w, http.StatusBadGateway, err.Error())
		return
	}
	s.audit(practiceID, "user", userID, "calendar.event_created", "calendar_config", config.ID, map[string]string{"summary": req.Summary})
	httpx.JSON(w, http.StatusCreated, event)
}

func (s *Server) exchangeGoogleCode(r *http.Request, code string) (googleTokenResponse, error) {
	values := url.Values{}
	values.Set("code", code)
	values.Set("client_id", s.cfg.GoogleClientID)
	values.Set("client_secret", s.cfg.GoogleClientSecret)
	values.Set("redirect_uri", s.cfg.GoogleRedirectURL)
	values.Set("grant_type", "authorization_code")
	return s.googleTokenRequest(r, values)
}

func (s *Server) refreshGoogleAccessToken(r *http.Request, config domain.CalendarConfig) (string, domain.CalendarConfig, error) {
	if config.OAuthRefreshTokenEnc == "" {
		return "", config, fmt.Errorf("google refresh token is missing")
	}
	refreshToken, err := s.decryptCalendarToken(config.OAuthRefreshTokenEnc)
	if err != nil {
		return "", config, fmt.Errorf("could not read refresh token")
	}
	values := url.Values{}
	values.Set("client_id", s.cfg.GoogleClientID)
	values.Set("client_secret", s.cfg.GoogleClientSecret)
	values.Set("refresh_token", refreshToken)
	values.Set("grant_type", "refresh_token")
	token, err := s.googleTokenRequest(r, values)
	if err != nil {
		return "", config, err
	}
	accessEnc, err := s.encryptCalendarToken(token.AccessToken)
	if err != nil {
		return "", config, fmt.Errorf("could not protect refreshed token")
	}
	expiresAt := time.Now().UTC().Add(time.Duration(token.ExpiresIn) * time.Second)
	config, ok := s.store.SaveCalendarOAuth(config.PracticeID, accessEnc, "", expiresAt)
	if !ok {
		return "", config, fmt.Errorf("calendar config not found")
	}
	return token.AccessToken, config, nil
}

func (s *Server) googleTokenRequest(r *http.Request, values url.Values) (googleTokenResponse, error) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(values.Encode()))
	if err != nil {
		return googleTokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return googleTokenResponse{}, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	var token googleTokenResponse
	_ = json.Unmarshal(body, &token)
	if res.StatusCode >= 300 {
		if token.ErrorDesc != "" {
			return googleTokenResponse{}, errors.New(token.ErrorDesc)
		}
		return googleTokenResponse{}, errors.New("google token request failed")
	}
	return token, nil
}

func (s *Server) postGoogleCalendarEvent(r *http.Request, config domain.CalendarConfig, accessToken, summary, description, start, end, attendee string) (map[string]any, error) {
	payload := map[string]any{
		"summary":     summary,
		"description": description,
		"start": map[string]string{
			"dateTime": start,
			"timeZone": config.Timezone,
		},
		"end": map[string]string{
			"dateTime": end,
			"timeZone": config.Timezone,
		},
	}
	if strings.TrimSpace(attendee) != "" {
		payload["attendees"] = []map[string]string{{"email": strings.TrimSpace(attendee)}}
	}
	body, _ := json.Marshal(payload)
	endpoint := "https://www.googleapis.com/calendar/v3/calendars/" + url.PathEscape(config.CalendarID) + "/events"
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	responseBody, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("google calendar event request failed")
	}
	var out map[string]any
	if err := json.Unmarshal(responseBody, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Server) signCalendarState(practiceID, userID string) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	payload := strings.Join([]string{practiceID, userID, strconv.FormatInt(time.Now().UTC().Unix(), 10), hex.EncodeToString(nonce)}, "|")
	mac := hmac.New(sha256.New, []byte(s.cfg.SessionSecret))
	_, _ = mac.Write([]byte(payload))
	signed := payload + "|" + hex.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(signed)), nil
}

func (s *Server) verifyCalendarState(state string) (string, string, bool) {
	decoded, err := base64.RawURLEncoding.DecodeString(state)
	if err != nil {
		return "", "", false
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 5 {
		return "", "", false
	}
	payload := strings.Join(parts[:4], "|")
	mac := hmac.New(sha256.New, []byte(s.cfg.SessionSecret))
	_, _ = mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[4])) {
		return "", "", false
	}
	createdAt, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || time.Since(time.Unix(createdAt, 0)) > 15*time.Minute {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func (s *Server) encryptCalendarToken(token string) (string, error) {
	block, err := aes.NewCipher(calendarTokenKey(s.cfg.CalendarTokenSecret))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(token), nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func (s *Server) decryptCalendarToken(encrypted string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(calendarTokenKey(s.cfg.CalendarTokenSecret))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", fmt.Errorf("invalid encrypted token")
	}
	nonce := raw[:gcm.NonceSize()]
	ciphertext := raw[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func calendarTokenKey(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}
