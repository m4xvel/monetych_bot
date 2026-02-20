package telegram

import (
	"context"
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
	h.answerCallback(cb, "")

	logger.Log.Infow("expert confirm order reaffirm initiated",
		"chat_id", chatID,
		"callback_data", cb.Data,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) != 2 {
		logger.Log.Warnw("invalid confirmed reaffirm callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	tokenCallback := parts[1]

	var payload ConfirmedAndDeclinedOrderSelectPayload

	if err := h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"confirmed_reaffirm",
		&payload,
	); err != nil {
		if isInvalidToken(err) {
			logger.Log.Warnw("invalid confirmed reaffirm callback token",
				"chat_id", chatID,
				"data", cb.Data,
				"err", err,
			)
			return
		}
		logger.Log.Errorw("failed to consume confirmed reaffirm callback token",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	orderID := payload.OrderID
	topicID := payload.TopicID
	threadID := payload.ThreadID

	if err := h.callbackTokenService.DeleteByActionAndOrderID(
		ctx,
		"back",
		orderID,
	); err != nil {
		logger.Log.Errorw("failed to delete back callbacks",
			"order_id", orderID,
			"err", err,
		)
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
		if isOrderAlreadyProcessed(err) {
			logger.Log.Infow("order already processed on expert confirm",
				"chat_id", chatID,
				"order_id", orderID,
				"err", err,
			)
			return
		}
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
		wrapped := wrapTelegramErr("telegram.send_expert_confirmed", err)
		logger.Log.Errorw("failed to send expert confirmation message",
			"chat_id", chatID,
			"order_id", orderID,
			"err", wrapped,
		)
	}

	if _, err := h.bot.Request(
		tgbotapi.NewDeleteTopicMessage(topicID, messageID, threadID),
	); err != nil {
		wrapped := wrapTelegramErr("telegram.delete_confirm_ui", err)
		logger.Log.Errorw("failed to delete confirmation ui message",
			"chat_id", chatID,
			"order_id", orderID,
			"err", wrapped,
		)
	}

	token, err := h.callbackTokenService.Create(
		ctx,
		"accept_client",
		&CancelOrderSelectPayload{
			ChatID:  chatID,
			OrderID: orderID,
		},
	)
	if err != nil {
		logger.Log.Errorw(
			"failed to create accept client order callback token",
			"err", err,
		)
	}

	clientMsg := tgbotapi.NewMessage(order.UserChatID, h.text.ConfirmYourOrder)
	btn := tgbotapi.NewInlineKeyboardButtonData(
		h.text.AcceptText,
		"accept_client:"+token,
	)
	clientMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn),
	)

	if _, err := h.bot.Send(clientMsg); err != nil {
		wrapped := wrapTelegramErr("telegram.notify_client_confirm", err)
		logger.Log.Errorw("failed to notify client about order confirmation",
			"order_id", orderID,
			"user_chat_id", order.UserChatID,
			"err", wrapped,
		)
		if err := h.callbackTokenService.Delete(
			ctx,
			token,
			"accept_client",
		); err != nil {
			logger.Log.Errorw("failed to cleanup accept client callback token",
				"order_id", orderID,
				"err", err,
			)
		}
	}
}
