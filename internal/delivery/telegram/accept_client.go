package telegram

import (
	"context"
	"fmt"
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
	h.answerCallback(cb, "")

	logger.Log.Infow("client confirm payment action initiated",
		"chat_id", chatID,
		"callback_data", cb.Data,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) != 2 {
		logger.Log.Warnw("invalid accept client callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	tokenCallback := parts[1]

	var payload CancelOrderSelectPayload

	if err := h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"accept_client",
		&payload,
	); err != nil {
		if isInvalidToken(err) {
			logger.Log.Warnw("invalid accept client callback token",
				"chat_id", chatID,
				"data", cb.Data,
				"err", err,
			)
			return
		}
		logger.Log.Errorw("failed to consume accept client callback token",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	orderID := payload.OrderID

	if err := h.orderService.SetCompletedStatus(ctx, orderID, chatID); err != nil {
		if isOrderAlreadyProcessed(err) {
			logger.Log.Infow("order already processed on client completion",
				"chat_id", chatID,
				"order_id", orderID,
				"err", err,
			)
			return
		}
		logger.Log.Errorw("failed to complete order by client",
			"chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
		return
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
			wrapped := wrapTelegramErr("telegram.notify_expert_completion", err)
			logger.Log.Errorw("failed to notify expert about order completion",
				"order_id", orderID,
				"topic_id", *order.TopicID,
				"err", wrapped,
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
		wrapped := wrapTelegramErr("telegram.edit_client_confirmation", err)
		logger.Log.Errorw("failed to edit client confirmation message",
			"chat_id", chatID,
			"order_id", orderID,
			"err", wrapped,
		)
	}

	buttons := make([]tgbotapi.InlineKeyboardButton, 0, 5)

	for i := 1; i <= 5; i++ {
		token, err := h.callbackTokenService.Create(
			ctx,
			"rate",
			&RateSelectPayload{
				ChatID:  chatID,
				Rate:    i,
				OrderID: orderID,
			},
		)
		if err != nil {
			continue
		}

		buttons = append(buttons,
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("⭐ %d", i),
				"rate:"+token,
			),
		)
	}

	rateMsg := tgbotapi.NewMessage(chatID, h.text.ChatClosedText)
	rateMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttons...),
	)

	if _, err := h.bot.Send(rateMsg); err != nil {
		wrapped := wrapTelegramErr("telegram.send_rate_prompt", err)
		logger.Log.Errorw("failed to send rate prompt to client",
			"chat_id", chatID,
			"order_id", orderID,
			"err", wrapped,
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
