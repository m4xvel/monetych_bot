package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SentOrder struct {
	ChatID    int64
	MessageID int
}

var orderMessages = map[int][]SentOrder{}

func (h *Handler) handleOrderSelect(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	user, _ := h.userService.GetUserByUserTgID(ctx, chatID)
	orderNew, _ := h.orderService.GetActiveByClient(ctx, user.ID, "new")
	orderActive, _ := h.orderService.GetActiveByClient(ctx, user.ID, "active")
	if orderNew != nil || orderActive != nil {
		h.bot.Send(tgbotapi.NewEditMessageText(
			chatID,
			messageID,
			h.text.AlreadyActiveOrder,
		))
		return
	}

	if !h.shouldProcess(chatID, messageID) {
		return
	}

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 3 {
		return
	}

	itemGame := parts[1]
	itemType := parts[2]

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	userID, _ := h.userService.GetUserByUserTgID(ctx, chatID)
	id, _ := h.orderService.CreateOrder(ctx, userID.ID)

	h.notifyAssessorsAboutOrder(ctx, id, itemGame, itemType, chatID, messageID)

	editText := tgbotapi.NewEditMessageText(
		chatID,
		cb.Message.MessageID,
		h.text.WaitingAssessor,
	)
	h.bot.Request(editText)

}

func (h *Handler) notifyAssessorsAboutOrder(
	ctx context.Context, orderID int, nameGame, nameType string, userID int64, messageUserId int) {
	tgIDs, err := h.assessorService.GetAllAssessorTgIDs(ctx)
	if err != nil {
		log.Printf("failed to get assessors: %v", err)
		return
	}
	for _, tgID := range tgIDs {
		msg := tgbotapi.NewMessage(tgID, h.textDynamic.NewOrder(orderID, nameGame, nameType))
		button := tgbotapi.NewInlineKeyboardButtonData(
			"Принять",
			fmt.Sprintf("accept:%d:%s:%s:%d:%d", orderID, nameGame, nameType, userID, messageUserId),
		)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(button),
		)
		sentMsg, _ := h.bot.Send(msg)
		orderMessages[orderID] = append(orderMessages[orderID], SentOrder{
			ChatID:    sentMsg.Chat.ID,
			MessageID: sentMsg.MessageID,
		})
	}
}
