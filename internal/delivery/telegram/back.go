package telegram

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleBack(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 4 {
		return
	}

	orderID, _ := strconv.Atoi(parts[1])
	topicID, _ := strconv.ParseInt(parts[2], 10, 64)
	threadID, _ := strconv.ParseInt(parts[3], 10, 64)

	order, err := h.orderService.GetOrderByID(ctx, orderID)
	if err != nil {
		return
	}

	h.renderEditControlPanel(
		cb.Message.MessageID,
		topicID,
		threadID,
		order,
	)
}
