package telegram

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handlerCancelSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	logger.Log.Infow("order cancel initiated",
		"chat_id", chatID,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 2 {
		logger.Log.Warnw("invalid cancel callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	orderID, err := strconv.Atoi(parts[1])
	if err != nil {
		logger.Log.Warnw("failed to parse order id for cancel",
			"chat_id", chatID,
			"value", parts[1],
		)
		return
	}

	logger.Log.Infow("deleting order messages before cancel",
		"order_id", orderID,
	)
	h.deleteOrderMessage(ctx, orderID)

	if err := h.orderService.SetCancelStatus(ctx, orderID); err != nil {
		logger.Log.Warnw("failed to cancel order",
			"order_id", orderID,
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	logger.Log.Infow("order cancelled",
		"order_id", orderID,
		"chat_id", chatID,
	)

	editText := tgbotapi.NewEditMessageText(
		chatID,
		cb.Message.MessageID,
		h.text.YouHaveCancelledOrder,
	)
	h.bot.Request(editText)
}
