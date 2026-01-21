package telegram

import (
	"context"
	"strconv"
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
	if len(parts) < 4 {
		logger.Log.Warnw("invalid declined reaffirm callback data",
			"data", cb.Data,
		)
		return
	}

	orderID, err := strconv.Atoi(parts[1])
	if err != nil {
		logger.Log.Warnw("failed to parse order id for decline reaffirm",
			"value", parts[1],
		)
		return
	}

	topicID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		logger.Log.Warnw("failed to parse topic id for decline reaffirm",
			"value", parts[2],
		)
		return
	}

	threadID, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		logger.Log.Warnw("failed to parse thread id for decline reaffirm",
			"value", parts[3],
		)
		return
	}

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
