package telegram

import (
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
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

	logger.Log.Infow("telegram bot authorized",
		"username", bot.Self.UserName,
		"debug", debug,
	)

	return bot, nil
}
