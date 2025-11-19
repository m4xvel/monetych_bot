package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleAcceptSelect(ctx context.Context, cb *tgbotapi.CallbackQuery) {
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
	messageUserID, _ := strconv.Atoi(parts[5])

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	h.deleteSentOrders(orderID)

	h.bot.Send(tgbotapi.NewMessage(
		chatID,
		h.textDynamic.AssessorAcceptedOrder(orderID, itemGame, itemType),
	))

	assessor, err := h.assessorService.GetByTgID(ctx, chatID)
	if err != nil {
		log.Printf("failed to get assessor: %v", err)
		return
	}

	threadID, err := h.createForumTopic(ctx, h.textDynamic.TitleOrderTopic(orderID, itemGame, itemType), chatID)
	if err != nil {
		log.Printf("failed to create forum topic: %v", err)
		return
	}

	h.orderService.AcceptOrder(ctx, assessor.ID, orderID, assessor.TopicID, threadID)
	h.sendOrderControlPanel(assessor.TopicID, threadID, orderID)

	h.bot.Request(tgbotapi.NewDeleteMessage(userID, messageUserID))
	h.bot.Send(tgbotapi.NewMessage(userID, h.text.AssessorAcceptedYourOrder))
}

func (h *Handler) sendOrderControlPanel(topicID int64, threadID int64, orderID int) {
	btnAccept := tgbotapi.NewInlineKeyboardButtonData(h.text.AcceptText, fmt.Sprintf("order_accept:%d", orderID))
	btnDecline := tgbotapi.NewInlineKeyboardButtonData(h.text.DeclineText, fmt.Sprintf("order_decline:%d", orderID))

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnAccept, btnDecline),
	)

	msg := tgbotapi.NewMessage(topicID, h.text.ApplicationManagementText)
	msg.MessageThreadID = threadID
	msg.ReplyMarkup = markup

	h.bot.Send(msg)
}
