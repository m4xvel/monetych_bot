package telegram

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleDeclinedReaffirmSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	messageID := cb.Message.MessageID

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))
	parts := strings.Split(cb.Data, ":")
	if len(parts) < 4 {
		return
	}
	orderID, _ := strconv.Atoi(parts[1])
	topicID, _ := strconv.ParseInt(parts[2], 10, 64)
	threadID, _ := strconv.ParseInt(parts[3], 10, 64)

	order, _ := h.orderService.GetOrderByID(ctx, orderID)

	h.orderService.SetDeclinedStatus(ctx, orderID)

	msg := tgbotapi.NewMessage(topicID, h.text.YouHaveCancelledOrder)
	msg.MessageThreadID = threadID
	h.bot.Send(msg)

	h.bot.Request(tgbotapi.NewDeleteTopicMessage(topicID, messageID, threadID))
	h.bot.Send(tgbotapi.NewMessage(order.UserChatID, h.text.YouOrderCancelled))

	h.stateService.SetStateIdle(ctx, order.UserChatID)
}
