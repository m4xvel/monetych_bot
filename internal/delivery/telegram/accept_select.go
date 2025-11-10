package telegram

import (
	"context"
	"fmt"
	"log"
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

	sentOrders := orderMessages[orderID]
	for _, sent := range sentOrders {
		deleteMsg := tgbotapi.NewDeleteMessage(sent.ChatID, sent.MessageID)
		if _, err := h.bot.Request(deleteMsg); err != nil {
			log.Println("Ошибка удаления сообщения:", err)
		}
	}

	delete(orderMessages, orderID)

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
		"Вы приняли заявку #%d ✅\n(%s, %s)",
		orderID, itemGame, itemType),
	)
	h.bot.Send(msg)

	h.createForumTopic(
		ctx,
		fmt.Sprintf("💼 Сделка #%d - (%s, %s)", orderID, itemGame, itemType),
		chatID,
	)

	editTextUser := tgbotapi.NewEditMessageText(
		userID,
		messageUserId,
		"✅ Оценщик принял Вашу заявку, продолжайте общаться в этом чате!",
	)
	_, _ = h.bot.Request(editTextUser)

}
