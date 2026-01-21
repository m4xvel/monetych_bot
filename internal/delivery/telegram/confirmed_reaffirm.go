package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleConfirmedReaffirmSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID
	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	logger.Log.Infow("expert confirm order reaffirm initiated",
		"chat_id", chatID,
		"callback_data", cb.Data,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 4 {
		logger.Log.Warnw("invalid confirmed reaffirm callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	orderID, err := strconv.Atoi(parts[1])
	if err != nil {
		logger.Log.Warnw("failed to parse order_id from confirmed reaffirm callback",
			"chat_id", chatID,
			"value", parts[1],
		)
		return
	}

	topicID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		logger.Log.Warnw("failed to parse topic_id from confirmed reaffirm callback",
			"chat_id", chatID,
			"value", parts[2],
		)
		return
	}

	threadID, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		logger.Log.Warnw("failed to parse thread_id from confirmed reaffirm callback",
			"chat_id", chatID,
			"value", parts[3],
		)
		return
	}

	order, err := h.orderService.GetOrderByID(ctx, orderID)
	if err != nil {
		logger.Log.Errorw("failed to get order for confirmation",
			"chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
		return
	}

	if err := h.orderService.SetExpertConfirmedStatus(ctx, orderID); err != nil {
		logger.Log.Errorw("failed to confirm order by expert",
			"chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
		return
	}

	logger.Log.Infow("order confirmed by expert",
		"chat_id", chatID,
		"order_id", orderID,
	)

	msg := tgbotapi.NewMessage(topicID, h.text.YouConfirmedOrder)
	msg.MessageThreadID = threadID
	if _, err := h.bot.Send(msg); err != nil {
		logger.Log.Errorw("failed to send expert confirmation message",
			"chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
	}

	if _, err := h.bot.Request(
		tgbotapi.NewDeleteTopicMessage(topicID, messageID, threadID),
	); err != nil {
		logger.Log.Errorw("failed to delete confirmation ui message",
			"chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
	}

	clientMsg := tgbotapi.NewMessage(order.UserChatID, h.text.ConfirmYourOrder)
	btn := tgbotapi.NewInlineKeyboardButtonData(
		h.text.AcceptText,
		fmt.Sprintf("accept_client:%d", orderID),
	)
	clientMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn),
	)

	if _, err := h.bot.Send(clientMsg); err != nil {
		logger.Log.Errorw("failed to notify client about order confirmation",
			"order_id", orderID,
			"user_chat_id", order.UserChatID,
			"err", err,
		)
	}
}
