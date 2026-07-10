package config

import (
	"bufio"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv              string
	HTTPAddr            string
	WebBaseURL          string
	DatabaseURL         string
	StoreDriver         string
	SessionSecret       string
	VAPIWebhookSecret   string
	WorkerToken         string
	LogLevelName        string
	SMTPHost            string
	SMTPPort            int
	SMTPUser            string
	SMTPPass            string
	SMTPSecure          string
	SMTPFromEmail       string
	SMTPFromName        string
	GoogleClientID      string
	GoogleClientSecret  string
	GoogleRedirectURL   string
	CalendarTokenSecret string
	StripeSecretKey     string
	StripeWebhookSecret string
	StripePriceID       string
}

func Load() Config {
	loadDotEnv(".env")
	return Config{
		AppEnv:              env("APP_ENV", "development"),
		HTTPAddr:            httpAddr(),
		WebBaseURL:          env("APP_BASE_URL", "http://localhost:3000"),
		DatabaseURL:         env("DATABASE_URL", ""),
		StoreDriver:         env("STORE_DRIVER", "memory"),
		SessionSecret:       env("SESSION_SECRET", "dev-session-secret-change-me"),
		VAPIWebhookSecret:   env("VAPI_WEBHOOK_SECRET", "dev-vapi-secret"),
		WorkerToken:         env("WORKER_TOKEN", "dev-worker-token"),
		LogLevelName:        env("LOG_LEVEL", "info"),
		SMTPHost:            env("SMTP_HOST", ""),
		SMTPPort:            envInt("SMTP_PORT", 587),
		SMTPUser:            env("SMTP_USER", ""),
		SMTPPass:            env("SMTP_PASS", ""),
		SMTPSecure:          env("SMTP_SECURE", "tls"),
		SMTPFromEmail:       env("SMTP_FROM_EMAIL", ""),
		SMTPFromName:        env("SMTP_FROM_NAME", "DentalDesk AI"),
		GoogleClientID:      env("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:  env("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:   env("GOOGLE_REDIRECT_URL", env("API_BASE_URL", "http://localhost:8080")+"/v1/calendar/oauth/callback"),
		CalendarTokenSecret: env("CALENDAR_TOKEN_SECRET", env("SESSION_SECRET", "dev-session-secret-change-me")),
		StripeSecretKey:     env("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: env("STRIPE_WEBHOOK_SECRET", ""),
		StripePriceID:       env("STRIPE_PRICE_ID", ""),
	}
}

func httpAddr() string {
	if value := strings.TrimSpace(os.Getenv("HTTP_ADDR")); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("PORT")); value != "" {
		return ":" + value
	}
	return ":8080"
}

func (c Config) LogLevel() slog.Level {
	switch strings.ToLower(c.LogLevelName) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		_ = os.Setenv(key, value)
	}
}
