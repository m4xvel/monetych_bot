package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleOrderAcceptAssessor(
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
	order := *h.orderService.GetOrderByID(ctx, orderID)
	msg := tgbotapi.NewMessage(order.UserID, h.text.ConfirmYourOrder)
	verificationButton := tgbotapi.NewInlineKeyboardButtonData(h.text.AcceptText, fmt.Sprintf("order_accept_client:%d", orderID))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(verificationButton),
	)
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
	h.bot.Request(tgbotapi.NewDeleteTopicMessage(*order.TopicID, messageID, *order.ThreadID))
}
