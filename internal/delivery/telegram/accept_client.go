package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleAcceptClientSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID
	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	logger.Log.Infow("client confirm payment action initiated",
		"chat_id", chatID,
		"callback_data", cb.Data,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 2 {
		logger.Log.Warnw("invalid accept client callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	orderID, err := strconv.Atoi(parts[1])
	if err != nil {
		logger.Log.Warnw("failed to parse order_id from accept client callback",
			"chat_id", chatID,
			"value", parts[1],
		)
		return
	}

	if err := h.orderService.SetCompletedStatus(ctx, orderID, chatID); err != nil {
		logger.Log.Errorw("failed to complete order by client",
			"chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
	}

	logger.Log.Infow("order completed by client",
		"chat_id", chatID,
		"order_id", orderID,
	)

	order, err := h.orderService.GetOrderByID(ctx, orderID)
	if err != nil {
		logger.Log.Errorw("failed to get order after completion",
			"chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
		return
	}

	if order.TopicID != nil && order.ThreadID != nil {
		msg := tgbotapi.NewMessage(*order.TopicID, h.text.OrderConfirmed)
		msg.MessageThreadID = *order.ThreadID

		if _, err := h.bot.Send(msg); err != nil {
			logger.Log.Errorw("failed to notify expert about order completion",
				"order_id", orderID,
				"topic_id", *order.TopicID,
				"err", err,
			)
		}
	}

	if _, err := h.bot.Request(
		tgbotapi.NewEditMessageText(
			chatID,
			messageID,
			h.text.YouConfirmedPayment,
		),
	); err != nil {
		logger.Log.Errorw("failed to edit client confirmation message",
			"chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
	}

	rateMsg := tgbotapi.NewMessage(chatID, h.text.ChatClosedText)
	rateMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⭐ 1", fmt.Sprintf("rate:%d:%d", 1, orderID)),
			tgbotapi.NewInlineKeyboardButtonData("⭐ 2", fmt.Sprintf("rate:%d:%d", 2, orderID)),
			tgbotapi.NewInlineKeyboardButtonData("⭐ 3", fmt.Sprintf("rate:%d:%d", 3, orderID)),
			tgbotapi.NewInlineKeyboardButtonData("⭐ 4", fmt.Sprintf("rate:%d:%d", 4, orderID)),
			tgbotapi.NewInlineKeyboardButtonData("⭐ 5", fmt.Sprintf("rate:%d:%d", 5, orderID)),
		),
	)

	if _, err := h.bot.Send(rateMsg); err != nil {
		logger.Log.Errorw("failed to send rate prompt to client",
			"chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
	}

	if err := h.stateService.SetStateIdle(ctx, chatID); err != nil {
		logger.Log.Errorw("failed to set idle state after order completion",
			"chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
	}
}
