package telegram

import (
	"context"
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
	if len(parts) != 2 {
		logger.Log.Warnw("invalid cancel callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	tokenCallback := parts[1]

	var payload CancelOrderSelectPayload

	h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"cancel",
		&payload,
	)

	if payload.ChatID != cb.From.ID {
		return
	}

	orderID := payload.OrderID

	h.callbackTokenService.DeleteByActionAndOrderID(
		ctx,
		"accept",
		orderID,
	)

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
