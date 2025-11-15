package telegram

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleOrderAcceptClient(
	ctx context.Context, cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	if !h.shouldProcess(chatID, messageID) {
		return
	}

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 2 {
		return
	}
	orderID, _ := strconv.Atoi(parts[1])
	h.orderService.UpdateStatus(ctx, orderID, "completed")
	order := h.orderService.GetOrderByID(ctx, orderID)

	msg := tgbotapi.NewMessage(
		*order.TopicID,
		h.text.OrderConfirmed,
	)
	msg.MessageThreadID = *order.ThreadID
	h.bot.Send(msg)
	editText := tgbotapi.NewEditMessageText(
		chatID,
		messageID,
		h.text.YouConfirmedPayment,
	)
	h.bot.Request(editText)
}
