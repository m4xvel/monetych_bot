package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                   string
	BotToken              string
	DatabaseURL           string
	KeyBase64             string
	Debug                 bool
	VerificationEnabled   bool
	PrivacyPolicyURL      string
	PublicOfferURL        string
	OrderMsgRetentionDays int
	WebhookEnabled        bool
	WebhookURL            string
	WebhookListenAddr     string
	WebhookDropPending    bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env:                   getEnv("APP_ENV", "dev"),
		BotToken:              os.Getenv("BOT_TOKEN"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		KeyBase64:             os.Getenv("CHAT_CRYPTO_KEY"),
		Debug:                 os.Getenv("DEBUG") == "true",
		VerificationEnabled:   getEnvBool("ENABLE_VERIFICATION", true),
		PrivacyPolicyURL:      os.Getenv("PRIVACY_POLICY_URL"),
		PublicOfferURL:        os.Getenv("PUBLIC_OFFER_URL"),
		OrderMsgRetentionDays: getEnvInt("ORDER_MESSAGES_RETENTION_DAYS", 30),
		WebhookEnabled:        getEnvBool("TELEGRAM_WEBHOOK_ENABLED", getEnv("APP_ENV", "dev") == "prod"),
		WebhookURL:            os.Getenv("TELEGRAM_WEBHOOK_URL"),
		WebhookListenAddr:     getEnv("TELEGRAM_WEBHOOK_LISTEN_ADDR", ":8080"),
		WebhookDropPending:    getEnvBool("TELEGRAM_WEBHOOK_DROP_PENDING_UPDATES", false),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.BotToken == "" {
		return fmt.Errorf("BOT_TOKEN is not set")
	}

	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	if c.KeyBase64 == "" {
		return fmt.Errorf("CHAT_CRYPTO_KEY is not set")
	}

	if c.PrivacyPolicyURL == "" {
		return fmt.Errorf("PRIVACY_POLICY_URL is not set")
	}

	if c.PublicOfferURL == "" {
		return fmt.Errorf("PUBLIC_OFFER_URL is not set")
	}

	if c.Env != "dev" && c.Env != "prod" {
		return fmt.Errorf("invalid APP_ENV: %s", c.Env)
	}

	if c.WebhookEnabled && c.WebhookURL == "" {
		return fmt.Errorf("TELEGRAM_WEBHOOK_URL is not set while TELEGRAM_WEBHOOK_ENABLED=true")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}

	return n
}

func getEnvBool(key string, defaultValue bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}

	switch v {
	case "1", "true", "TRUE", "yes", "YES":
		return true
	case "0", "false", "FALSE", "no", "NO":
		return false
	default:
		return defaultValue
	}
}
