package telegram

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleBack(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	h.answerCallback(cb, "")

	logger.Log.Infow("back navigation action initiated",
		"chat_id", chatID,
		"callback_data", cb.Data,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) != 2 {
		logger.Log.Warnw("invalid back callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	tokenCallback := parts[1]

	var payload ConfirmedAndDeclinedOrderSelectPayload

	if err := h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"back",
		&payload,
	); err != nil {
		if isInvalidToken(err) {
			logger.Log.Warnw("invalid back callback token",
				"chat_id", chatID,
				"data", cb.Data,
				"err", err,
			)
			return
		}
		logger.Log.Errorw("failed to consume back callback token",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	orderID := payload.OrderID
	topicID := payload.TopicID
	threadID := payload.ThreadID

	if err := h.callbackTokenService.DeleteByActionAndOrderID(
		ctx,
		"declined_reaffirm",
		orderID,
	); err != nil {
		logger.Log.Errorw("failed to delete declined reaffirm callbacks",
			"order_id", orderID,
			"err", err,
		)
	}

	if err := h.callbackTokenService.DeleteByActionAndOrderID(
		ctx,
		"confirmed_reaffirm",
		orderID,
	); err != nil {
		logger.Log.Errorw("failed to delete confirmed reaffirm callbacks",
			"order_id", orderID,
			"err", err,
		)
	}

	order, err := h.orderService.GetOrderByID(ctx, orderID)
	if err != nil {
		logger.Log.Errorw("failed to get order for back navigation",
			"chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
		return
	}

	h.renderEditControlPanel(
		ctx,
		cb.Message.MessageID,
		topicID,
		threadID,
		order,
	)
}
