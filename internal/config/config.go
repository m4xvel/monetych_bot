package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	BotToken    string
	DatabaseURL string
	Debug       bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env:         getEnv("APP_ENV", "dev"),
		BotToken:    os.Getenv("BOT_TOKEN"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Debug:       os.Getenv("DEBUG") == "true",
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

	if c.Env != "dev" && c.Env != "prod" {
		return fmt.Errorf("invalid APP_ENV: %s", c.Env)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
