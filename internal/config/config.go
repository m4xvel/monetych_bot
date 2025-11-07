package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken    string
	DatabaseURL string
	Debug       bool
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env file not found, using system environment variables")
	}

	cfg := &Config{
		BotToken:    os.Getenv("BOT_TOKEN"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Debug:       os.Getenv("DEBUG") == "true",
	}

	if cfg.BotToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN is not set")
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	return cfg, nil
}
