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

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	sentOrders := orderMessages[orderID]
	for _, sent := range sentOrders {
		deleteMsg := tgbotapi.NewDeleteMessage(sent.ChatID, sent.MessageID)
		h.bot.Request(deleteMsg)
	}

	delete(orderMessages, orderID)

	msg := tgbotapi.NewMessage(
		chatID,
		h.textDynamic.AssessorAcceptedOrder(orderID, itemGame, itemType),
	)
	h.bot.Send(msg)

	threadID, _ := h.createForumTopic(
		ctx,
		h.textDynamic.TitleOrderTopic(orderID, itemGame, itemType),
		chatID,
	)
	topicID := h.assessorService.GetTopicIDByTgID(ctx, chatID)
	h.orderService.Accept(ctx, chatID, orderID, topicID, threadID)
	h.sendOrderControlPanel(topicID, threadID, orderID)
	h.bot.Request(tgbotapi.NewDeleteMessage(userID, messageUserId))
	h.bot.Send(tgbotapi.NewMessage(
		userID,
		h.text.AssessorAcceptedYourOrder,
	))
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
