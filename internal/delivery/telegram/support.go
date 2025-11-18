package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleSupportCommand(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	h.bot.Send(tgbotapi.NewMessage(
		chatID,
		h.text.SupportText,
	))
}
