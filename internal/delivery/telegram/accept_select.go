package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleAcceptSelect(
	ctx context.Context, cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	if !h.shouldProcess(chatID, messageID) {
		return
	}

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 6 {
		return
	}
	orderID, _ := strconv.Atoi(parts[1])
	itemGame := parts[2]
	itemType := parts[3]
	userID, _ := strconv.ParseInt(parts[4], 10, 64)
	messageUserId, _ := strconv.Atoi(parts[5])

	_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	h.orderService.Accept(ctx, chatID, orderID)

	editText := tgbotapi.NewEditMessageText(
		chatID,
		cb.Message.MessageID,
		fmt.Sprintf("Вы приняли заявку #%d ✅ (%s, %s)",
			orderID, itemGame, itemType),
	)
	_, _ = h.bot.Request(editText)

	h.bot.Send(tgbotapi.NewMessage(chatID,
		"Перейдите в чат, чтобы продолжить..."),
	)

	editTextUser := tgbotapi.NewEditMessageText(
		userID,
		messageUserId,
		"Оценщик принял Вашу заявку ✅\nПерейдите в чат, чтобы продолжить...",
	)
	_, _ = h.bot.Request(editTextUser)

}
