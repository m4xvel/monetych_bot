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
	h.answerCallback(cb, "")

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

	if err := h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"declined_reaffirm",
		&payload,
	); err != nil {
		if isInvalidToken(err) {
			logger.Log.Warnw("invalid declined reaffirm callback token",
				"data", cb.Data,
				"err", err,
			)
			return
		}
		logger.Log.Errorw("failed to consume declined reaffirm callback token",
			"err", err,
		)
		return
	}

	orderID := payload.OrderID
	topicID := payload.TopicID
	threadID := payload.ThreadID

	if err := h.callbackTokenService.DeleteByActionAndOrderID(
		ctx,
		"back",
		orderID,
	); err != nil {
		logger.Log.Errorw("failed to delete back callbacks",
			"order_id", orderID,
			"err", err,
		)
	}

	order, err := h.orderService.GetOrderByID(ctx, orderID)
	if err != nil || order == nil {
		logger.Log.Errorw("failed to get order for decline reaffirm",
			"order_id", orderID,
		)
		return
	}

	if err := h.orderService.SetDeclinedStatus(ctx, orderID); err != nil {
		if isOrderAlreadyProcessed(err) {
			logger.Log.Infow("order already processed on decline",
				"order_id", orderID,
				"err", err,
			)
			return
		}
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
	if _, err := h.bot.Send(msg); err != nil {
		wrapped := wrapTelegramErr("telegram.send_decline_notice", err)
		logger.Log.Errorw("failed to send decline notice",
			"order_id", orderID,
			"topic_id", topicID,
			"err", wrapped,
		)
	}

	if _, err := h.bot.Request(tgbotapi.NewDeleteTopicMessage(topicID, messageID, threadID)); err != nil {
		wrapped := wrapTelegramErr("telegram.delete_decline_ui", err)
		logger.Log.Errorw("failed to delete decline confirmation message",
			"order_id", orderID,
			"topic_id", topicID,
			"err", wrapped,
		)
	}
	if _, err := h.bot.Send(tgbotapi.NewMessage(order.UserChatID, h.text.YouOrderCancelled)); err != nil {
		wrapped := wrapTelegramErr("telegram.notify_user_declined", err)
		logger.Log.Errorw("failed to notify user about declined order",
			"order_id", orderID,
			"user_chat_id", order.UserChatID,
			"err", wrapped,
		)
	}

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
