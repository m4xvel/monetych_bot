package telegram

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handlerCancelSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))
	parts := strings.Split(cb.Data, ":")
	if len(parts) < 2 {
		return
	}

	orderID, _ := strconv.Atoi(parts[1])

	h.deleteOrderMessage(ctx, orderID)
	err := h.orderService.SetCancelStatus(ctx, orderID)
	if err != nil {
		return
	}

	editText := tgbotapi.NewEditMessageText(
		chatID,
		cb.Message.MessageID,
		h.text.YouHaveCancelledOrder,
	)
	h.bot.Request(editText)
}
