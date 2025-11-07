package telegram

import (
	"errors"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func NewBot(token string, debug bool) (*tgbotapi.BotAPI, error) {
	if token == "" {
		return nil, errors.New("telegram bot token is empty")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Telegram bot: %w", err)
	}

	bot.Debug = debug
	log.Printf("Authorized on account %s (debug: %v)", bot.Self.UserName, debug)
	return bot, nil
}
