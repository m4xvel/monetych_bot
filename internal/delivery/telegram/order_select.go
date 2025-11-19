package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleOrderSelect(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

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

	userID, _ := h.userService.GetByTgID(ctx, chatID)
	id, err := h.orderService.CreateOrder(ctx, userID.ID)
	if id == 0 {
		h.bot.Send(tgbotapi.NewEditMessageText(
			chatID,
			messageID,
			h.text.AlreadyActiveOrder,
		))
		return
	}
	if err != nil {
		log.Printf("failed to create order: %v", err)
		return
	}

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
	tgIDs, err := h.assessorService.GetAllTgIDs(ctx)
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

		sentMsg, err := h.bot.Send(msg)
		if err != nil {
			log.Printf("failed to send order to assessor %d: %v", tgID, err)
			continue
		}

		h.addSentOrder(orderID, SentOrder{
			ChatID:    sentMsg.Chat.ID,
			MessageID: sentMsg.MessageID,
		})
	}
}
