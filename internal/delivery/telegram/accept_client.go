package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleAcceptClientSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))
	parts := strings.Split(cb.Data, ":")
	if len(parts) < 2 {
		return
	}
	orderID, _ := strconv.Atoi(parts[1])

	h.orderService.SetCompletedStatus(ctx, orderID)

	order, _ := h.orderService.GetOrderByID(ctx, orderID)

	msg := tgbotapi.NewMessage(*order.TopicID, h.text.OrderConfirmed)
	msg.MessageThreadID = *order.ThreadID
	h.bot.Send(msg)

	h.bot.Request(tgbotapi.NewEditMessageText(
		chatID,
		messageID,
		h.text.YouConfirmedPayment,
	))

	msg = tgbotapi.NewMessage(chatID, h.text.ChatClosedText)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⭐ 1", fmt.Sprintf("rate:%d:%d", 1, orderID)),
			tgbotapi.NewInlineKeyboardButtonData("⭐ 2", fmt.Sprintf("rate:%d:%d", 2, orderID)),
			tgbotapi.NewInlineKeyboardButtonData("⭐ 3", fmt.Sprintf("rate:%d:%d", 3, orderID)),
			tgbotapi.NewInlineKeyboardButtonData("⭐ 4", fmt.Sprintf("rate:%d:%d", 4, orderID)),
			tgbotapi.NewInlineKeyboardButtonData("⭐ 5", fmt.Sprintf("rate:%d:%d", 5, orderID)),
		),
	)
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)

	h.stateService.SetStateIdle(ctx, chatID)
}
