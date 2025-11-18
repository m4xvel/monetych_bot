package telegram

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleOrderDeclineAssessor(
	ctx context.Context, cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	if !h.shouldProcess(chatID, messageID) {
		return
	}

	_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 2 {
		return
	}
	orderID, _ := strconv.Atoi(parts[1])
	order := *h.orderService.GetOrderByID(ctx, orderID)
	h.orderService.UpdateStatus(ctx, orderID, "canceled")

	msgAssessor := tgbotapi.NewMessage(
		*order.TopicID,
		h.text.YouHaveCancelledOrder,
	)
	msgAssessor.MessageThreadID = *order.ThreadID
	h.bot.Send(msgAssessor)
	h.bot.Request(tgbotapi.NewDeleteTopicMessage(*order.TopicID, messageID, *order.ThreadID))
	user, _ := h.userService.GetUserByUserID(ctx, order.UserID)
	h.bot.Send(tgbotapi.NewMessage(user.UserID, h.text.YouOrderCancelled))
}
