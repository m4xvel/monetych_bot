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
	itemType := parts[2]

	editText := tgbotapi.NewEditMessageText(
		chatID,
		cb.Message.MessageID,
		fmt.Sprintf("📦 Вы выбрали: %s", itemType),
	)
	_, _ = h.bot.Request(editText)

	isVerified := h.userService.CheckStatusVerified(ctx, chatID)
	if !isVerified {
		h.showInlineKeyboardVerification(
			chatID,
			"Для безопасной сделки необходимо подтвердить вашу личность. Это просто и не займет много времени:",
			false,
		)
		return
	}

	h.contactAnAppraiser(chatID)
}
