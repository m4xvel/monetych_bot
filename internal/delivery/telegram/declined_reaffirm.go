package telegram

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleDeclinedReaffirmSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	messageID := cb.Message.MessageID
	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	logger.Log.Infow("decline reaffirm clicked",
		"callback_data", cb.Data,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) != 2 {
		logger.Log.Warnw("invalid declined reaffirm callback data",
			"data", cb.Data,
		)
		return
	}

	tokenCallback := parts[1]

	var payload ConfirmedAndDeclinedOrderSelectPayload

	h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"declined_reaffirm",
		&payload,
	)

	orderID := payload.OrderID
	topicID := payload.TopicID
	threadID := payload.ThreadID

	h.callbackTokenService.DeleteByActionAndOrderID(
		ctx,
		"back",
		orderID,
	)

	order, err := h.orderService.GetOrderByID(ctx, orderID)
	if err != nil || order == nil {
		logger.Log.Errorw("failed to get order for decline reaffirm",
			"order_id", orderID,
		)
		return
	}

	if err := h.orderService.SetDeclinedStatus(ctx, orderID); err != nil {
		logger.Log.Warnw("failed to decline order",
			"order_id", orderID,
			"err", err,
		)
		return
	}

	logger.Log.Infow("order declined",
		"order_id", orderID,
	)

	msg := tgbotapi.NewMessage(topicID, h.text.YouHaveCancelledOrder)
	msg.MessageThreadID = threadID
	h.bot.Send(msg)

	h.bot.Request(tgbotapi.NewDeleteTopicMessage(topicID, messageID, threadID))
	h.bot.Send(tgbotapi.NewMessage(order.UserChatID, h.text.YouOrderCancelled))

	if err := h.stateService.SetStateIdle(ctx, order.UserChatID); err != nil {
		logger.Log.Errorw("failed to set user idle after decline",
			"order_id", orderID,
			"user_chat_id", order.UserChatID,
			"err", err,
		)
		return
	}

	logger.Log.Infow("user returned to idle after decline",
		"order_id", orderID,
		"user_chat_id", order.UserChatID,
	)
}
