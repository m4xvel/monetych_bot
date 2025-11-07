package telegram

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleTypeSelect(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	if !h.shouldProcess(chatID, messageID) {
		return
	}

	_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 3 {
		return
	}
	gameID := parts[1]
	itemType := parts[2]

	editText := tgbotapi.NewEditMessageText(
		chatID,
		cb.Message.MessageID,
		fmt.Sprintf("📦 Вы выбрали: %s", itemType),
	)
	_, _ = h.bot.Request(editText)

	msg := fmt.Sprintf("Вы выбрали игру ID=%s и тип=%s ✅", gameID, itemType)
	h.bot.Send(tgbotapi.NewMessage(chatID, msg))
}
