package app

import (
	"crypto/hmac"
	"crypto/sha256"
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

type stripeCheckoutSession struct {
	ID           string            `json:"id"`
	URL          string            `json:"url"`
	Customer     string            `json:"customer"`
	Subscription string            `json:"subscription"`
	ClientRefID  string            `json:"client_reference_id"`
	Metadata     map[string]string `json:"metadata"`
}

type stripeWebhookEvent struct {
	Type string `json:"type"`
	Data struct {
		Object json.RawMessage `json:"object"`
	} `json:"data"`
}

func (s *Server) createStripeCheckoutSession(w http.ResponseWriter, r *http.Request, userID, practiceID string) {
	_ = userID
	if s.cfg.StripeSecretKey == "" || s.cfg.StripePriceID == "" {
		httpx.Error(w, http.StatusBadRequest, "stripe checkout is not configured")
		return
	}
	values := url.Values{}
	values.Set("mode", "subscription")
	values.Set("success_url", s.cfg.WebBaseURL+"/?billing=success")
	values.Set("cancel_url", s.cfg.WebBaseURL+"/?billing=cancelled")
	values.Set("client_reference_id", practiceID)
	values.Set("line_items[0][price]", s.cfg.StripePriceID)
	values.Set("line_items[0][quantity]", "1")
	values.Set("metadata[practice_id]", practiceID)

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, "https://api.stripe.com/v1/checkout/sessions", strings.NewReader(values.Encode()))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not create checkout request")
		return
	}
	req.SetBasicAuth(s.cfg.StripeSecretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		httpx.Error(w, http.StatusBadGateway, "stripe checkout request failed")
		return
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		httpx.Error(w, http.StatusBadGateway, "stripe checkout request failed")
		return
	}
	var session stripeCheckoutSession
	if err := json.Unmarshal(body, &session); err != nil {
		httpx.Error(w, http.StatusBadGateway, "invalid stripe checkout response")
		return
	}
	httpx.JSON(w, http.StatusCreated, session)
}

func (s *Server) stripeWebhook(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if s.cfg.StripeWebhookSecret != "" {
		if err := verifyStripeSignature(r.Header.Get("Stripe-Signature"), body, s.cfg.StripeWebhookSecret); err != nil {
			httpx.Error(w, http.StatusBadRequest, "invalid stripe signature")
			return
		}
	}
	var event stripeWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid stripe webhook")
		return
	}
	switch event.Type {
	case "checkout.session.completed":
		s.handleStripeCheckoutCompleted(event.Data.Object)
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "received"})
}

func (s *Server) handleStripeCheckoutCompleted(raw json.RawMessage) {
	var session stripeCheckoutSession
	if err := json.Unmarshal(raw, &session); err != nil {
		s.logger.Error("stripe checkout session parse failed", "error", err)
		return
	}
	practiceID := session.ClientRefID
	if practiceID == "" && session.Metadata != nil {
		practiceID = session.Metadata["practice_id"]
	}
	if practiceID == "" {
		s.logger.Error("stripe checkout session missing practice id", "session_id", session.ID)
		return
	}
	subscription, ok := s.store.GetBillingSubscription(practiceID)
	if !ok {
		subscription = domain.BillingSubscription{PracticeID: practiceID}
	}
	subscription.Status = "active"
	subscription.StripeCustomerID = session.Customer
	subscription.StripeSubscriptionID = session.Subscription
	saved := s.store.SaveBillingSubscription(subscription)
	s.audit(practiceID, "webhook", "stripe", "billing.stripe_checkout_completed", "billing_subscription", saved.ID, map[string]string{"sessionId": session.ID})
}

func verifyStripeSignature(header string, body []byte, secret string) error {
	timestamp := ""
	signatures := make([]string, 0)
	for _, part := range strings.Split(header, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "t":
			timestamp = value
		case "v1":
			signatures = append(signatures, value)
		}
	}
	if timestamp == "" || len(signatures) == 0 {
		return errors.New("missing stripe signature fields")
	}
	createdAt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return err
	}
	if time.Since(time.Unix(createdAt, 0)) > 5*time.Minute {
		return errors.New("stripe signature expired")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(fmt.Sprintf("%s.%s", timestamp, string(body))))
	expected := hex.EncodeToString(mac.Sum(nil))
	for _, signature := range signatures {
		if hmac.Equal([]byte(expected), []byte(signature)) {
			return nil
		}
	}
	return errors.New("stripe signature mismatch")
}
