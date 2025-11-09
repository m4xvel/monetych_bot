package telegram

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleOrderSelect(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	if !h.shouldProcess(chatID, messageID) {
		return
	}

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 3 {
		return
	}

	itemGame := parts[1]
	itemType := parts[2]

	_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	id, _ := h.orderService.CreateOrder(ctx, chatID)

	h.notifyAssessorsAboutOrder(ctx, id, itemGame, itemType, chatID, messageID)

	editText := tgbotapi.NewEditMessageText(
		chatID,
		cb.Message.MessageID,
		"Оценщик уже спешит к Вам ⏳",
	)
	_, _ = h.bot.Request(editText)

}
