package notifications

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"

	"dentaldesk/apps/api/internal/config"
	"dentaldesk/apps/api/internal/domain"
)

type Service struct {
	logger *slog.Logger
	cfg    config.Config
}

func NewService(logger *slog.Logger, cfg config.Config) *Service {
	return &Service{logger: logger, cfg: cfg}
}

func (s *Service) SendStaffSummary(ctx context.Context, practiceID, subject, body string) error {
	_ = ctx
	s.logger.Info("staff summary queued", "practice_id", practiceID, "subject", subject, "body_len", len(body))
	return nil
}

func (s *Service) SendCallSummaryEmail(ctx context.Context, toEmail string, summary domain.CallSummary) error {
	if strings.TrimSpace(toEmail) == "" {
		s.logger.Info("staff summary email skipped; no notification email", "practice_id", summary.PracticeID)
		return nil
	}
	subject := "New DentalDesk call summary"
	text := fmt.Sprintf(`Caller: %s
Reason: %s
Urgency: %s
AI action: %s
Follow-up needed: %s

Summary:
%s`, summary.CallerName, summary.Reason, summary.Urgency, summary.AIAction, summary.FollowUpNeeded, summary.Summary)
	html := fmt.Sprintf(
		`<h2>New DentalDesk call summary</h2><p><strong>Caller:</strong> %s</p><p><strong>Reason:</strong> %s</p><p><strong>Urgency:</strong> %s</p><p><strong>AI action:</strong> %s</p><p><strong>Follow-up needed:</strong> %s</p><p><strong>Summary:</strong><br>%s</p>`,
		escapeHTML(summary.CallerName),
		escapeHTML(summary.Reason),
		escapeHTML(summary.Urgency),
		escapeHTML(summary.AIAction),
		escapeHTML(summary.FollowUpNeeded),
		escapeHTML(summary.Summary),
	)
	return s.sendEmail(ctx, toEmail, subject, text, html)
}

func (s *Service) SendAppointmentRequestEmail(ctx context.Context, toEmail string, request domain.AppointmentRequest) error {
	if strings.TrimSpace(toEmail) == "" {
		s.logger.Info("appointment request email skipped; no notification email", "practice_id", request.PracticeID)
		return nil
	}
	subject := "New DentalDesk appointment request"
	text := fmt.Sprintf(`Caller: %s
Phone: %s
Request type: %s
Preferred time: %s
Insurance: %s
Status: %s

Notes:
%s`, request.CallerName, request.CallerPhone, request.RequestType, request.PreferredTime, request.Insurance, request.Status, request.Notes)
	html := fmt.Sprintf(
		`<h2>New DentalDesk appointment request</h2><p><strong>Caller:</strong> %s</p><p><strong>Phone:</strong> %s</p><p><strong>Request type:</strong> %s</p><p><strong>Preferred time:</strong> %s</p><p><strong>Insurance:</strong> %s</p><p><strong>Status:</strong> %s</p><p><strong>Notes:</strong><br>%s</p>`,
		escapeHTML(request.CallerName),
		escapeHTML(request.CallerPhone),
		escapeHTML(request.RequestType),
		escapeHTML(request.PreferredTime),
		escapeHTML(request.Insurance),
		escapeHTML(request.Status),
		escapeHTML(request.Notes),
	)
	return s.sendEmail(ctx, toEmail, subject, text, html)
}

func (s *Service) SendInviteEmail(ctx context.Context, toEmail, role, inviteURL string) error {
	subject := "You're invited to DentalDesk AI"
	text := fmt.Sprintf("You have been invited to DentalDesk AI as %s.\n\nAccept your invite:\n%s\n\nThis invite expires in 7 days.", role, inviteURL)
	html := fmt.Sprintf(
		`<p>You have been invited to DentalDesk AI as <strong>%s</strong>.</p><p><a href="%s">Accept your invite</a></p><p>This invite expires in 7 days.</p>`,
		escapeHTML(role),
		escapeHTML(inviteURL),
	)
	return s.sendEmail(ctx, toEmail, subject, text, html)
}

func (s *Service) SendEmailVerification(ctx context.Context, toEmail, verifyURL string) error {
	subject := "Verify your DentalDesk AI email"
	text := fmt.Sprintf("Verify your DentalDesk AI email address:\n%s\n\nThis link expires in 24 hours.", verifyURL)
	html := fmt.Sprintf(`<p>Verify your DentalDesk AI email address.</p><p><a href="%s">Verify email</a></p><p>This link expires in 24 hours.</p>`, escapeHTML(verifyURL))
	return s.sendEmail(ctx, toEmail, subject, text, html)
}

func (s *Service) SendPasswordReset(ctx context.Context, toEmail, resetURL string) error {
	subject := "Reset your DentalDesk AI password"
	text := fmt.Sprintf("Reset your DentalDesk AI password:\n%s\n\nThis link expires in 1 hour. If you did not request this, ignore this email.", resetURL)
	html := fmt.Sprintf(`<p>Reset your DentalDesk AI password.</p><p><a href="%s">Reset password</a></p><p>This link expires in 1 hour. If you did not request this, ignore this email.</p>`, escapeHTML(resetURL))
	return s.sendEmail(ctx, toEmail, subject, text, html)
}

func (s *Service) sendEmail(ctx context.Context, toEmail, subject, textBody, htmlBody string) error {
	_ = ctx
	if s.cfg.SMTPHost == "" || s.cfg.SMTPUser == "" || s.cfg.SMTPPass == "" {
		s.logger.Info("email dev fallback", "to", toEmail, "subject", subject, "body_len", len(textBody))
		return nil
	}

	fromEmail := s.cfg.SMTPFromEmail
	if fromEmail == "" {
		fromEmail = s.cfg.SMTPUser
	}
	from := mail.Address{Name: s.cfg.SMTPFromName, Address: fromEmail}
	to := mail.Address{Address: toEmail}
	message := buildMessage(from, to, subject, textBody, htmlBody)

	addr := net.JoinHostPort(s.cfg.SMTPHost, fmt.Sprint(s.cfg.SMTPPort))
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
	secure := strings.ToLower(strings.TrimSpace(s.cfg.SMTPSecure))
	if secure == "ssl" || secure == "smtps" || s.cfg.SMTPPort == 465 {
		return sendImplicitTLS(addr, s.cfg.SMTPHost, auth, from.Address, []string{to.Address}, []byte(message))
	}
	return sendStartTLS(addr, s.cfg.SMTPHost, auth, from.Address, []string{to.Address}, []byte(message))
}

func sendStartTLS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	return sendSMTPMessage(client, from, to, msg)
}

func sendImplicitTLS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	return sendSMTPMessage(client, from, to, msg)
}

func sendSMTPMessage(client *smtp.Client, from string, to []string, msg []byte) error {
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func buildMessage(from, to mail.Address, subject, textBody, htmlBody string) string {
	boundary := "dentaldesk-boundary"
	headers := []string{
		"From: " + from.String(),
		"To: " + to.String(),
		"Subject: " + mime.QEncoding.Encode("utf-8", subject),
		"MIME-Version: 1.0",
		`Content-Type: multipart/alternative; boundary="` + boundary + `"`,
	}
	return strings.Join(headers, "\r\n") + "\r\n\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n\r\n" +
		textBody + "\r\n\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n\r\n" +
		htmlBody + "\r\n\r\n" +
		"--" + boundary + "--\r\n"
}

func escapeHTML(value string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;")
	return replacer.Replace(value)
}
