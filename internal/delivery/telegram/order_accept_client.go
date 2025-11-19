package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/domain"
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
	h.orderService.UpdateStatus(ctx, orderID, domain.OrderCompleted)
	order, _ := h.orderService.GetOrderByID(ctx, orderID)

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

	msg = tgbotapi.NewMessage(chatID, "Чат закрыт! Оцените наш сервис от 1 до 5 ⭐")
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
}
