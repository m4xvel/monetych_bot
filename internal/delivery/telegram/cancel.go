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
	h.answerCallback(cb, "")

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

	if err := h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"cancel",
		&payload,
	); err != nil {
		if isInvalidToken(err) {
			logger.Log.Warnw("invalid cancel callback token",
				"chat_id", chatID,
				"data", cb.Data,
				"err", err,
			)
			return
		}
		logger.Log.Errorw("failed to consume cancel callback token",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	if payload.ChatID != cb.From.ID {
		return
	}

	orderID := payload.OrderID

	if err := h.callbackTokenService.DeleteByActionAndOrderID(
		ctx,
		"accept",
		orderID,
	); err != nil {
		logger.Log.Errorw("failed to delete accept callbacks",
			"order_id", orderID,
			"err", err,
		)
	}

	logger.Log.Infow("deleting order messages before cancel",
		"order_id", orderID,
	)
	h.deleteOrderMessage(ctx, orderID)

	if err := h.orderService.SetCancelStatus(ctx, orderID); err != nil {
		if isOrderAlreadyProcessed(err) {
			logger.Log.Infow("order already processed on cancel",
				"order_id", orderID,
				"chat_id", chatID,
				"err", err,
			)
			return
		}
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

	if _, err := h.bot.Request(editText); err != nil {
		wrapped := wrapTelegramErr("telegram.edit_cancel_message", err)
		logger.Log.Errorw("failed to edit cancel message",
			"order_id", orderID,
			"chat_id", chatID,
			"err", wrapped,
		)
	}
}
