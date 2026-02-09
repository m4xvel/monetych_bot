package telegram

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleVerificationSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	h.bot.Request(tgbotapi.NewCallback(cb.ID, "Отправлено пользователю"))

	logger.Log.Infow("verification request initiated",
		"chat_id", chatID,
		"callback_data", cb.Data,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) != 2 {
		logger.Log.Warnw("invalid verification callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	tokenCallback := parts[1]

	var payload VerificationSelectPayload

	if err := h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"verification",
		&payload,
	); err != nil {
		logger.Log.Errorw("failed to consume verification callback token",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	tokenVerify, err := h.callbackTokenService.Create(
		ctx,
		"verify",
		&payload,
	)
	if err != nil {
		logger.Log.Errorw("failed to create verify callback token",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	btn := tgbotapi.NewInlineKeyboardButtonData(
		h.text.VerifyButtonText,
		"verify:"+tokenVerify,
	)

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn),
	)

	message := tgbotapi.NewMessage(
		payload.UserChatID,
		h.text.YouNeedToVerify,
	)
	message.ReplyMarkup = &markup

	if _, err := h.bot.Send(message); err != nil {
		logger.Log.Errorw("failed to send verification message to user",
			"chat_id", payload.UserChatID,
			"order_id", payload.OrderID,
			"err", err,
		)
		return
	}
}

func (h *Handler) handleVerifySelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	h.bot.Request(tgbotapi.NewCallback(cb.ID, "Запрос получен"))

	logger.Log.Infow("user verification button clicked",
		"chat_id", chatID,
		"callback_data", cb.Data,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) != 2 {
		logger.Log.Warnw("invalid verify callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	tokenCallback := parts[1]

	var payload VerificationSelectPayload

	if err := h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"verify",
		&payload,
	); err != nil {
		logger.Log.Errorw("failed to consume verify callback token",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	if payload.UserChatID != cb.From.ID {
		logger.Log.Warnw("verify callback user mismatch",
			"chat_id", chatID,
			"user_chat_id", payload.UserChatID,
		)
		return
	}

	if err := h.userService.SetVerified(ctx, payload.UserChatID, true); err != nil {
		logger.Log.Errorw("failed to set user verified",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	h.callbackTokenService.DeleteByActionAndOrderID(
		ctx,
		"verification",
		payload.OrderID,
	)

	edit := tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      chatID,
			MessageID:   cb.Message.MessageID,
			ReplyMarkup: nil,
		},
	}

	if _, err := h.bot.Send(edit); err != nil {
		logger.Log.Errorw("failed to remove verification keyboard",
			"chat_id", chatID,
			"err", err,
		)
	}

	if _, err := h.bot.Send(tgbotapi.NewMessage(
		chatID,
		h.text.SuccessfulVerify,
	)); err != nil {
		logger.Log.Errorw("failed to send verify success message",
			"chat_id", chatID,
			"err", err,
		)
	}

	order, err := h.orderService.GetOrderByID(ctx, payload.OrderID)
	if err != nil || order == nil {
		logger.Log.Errorw("failed to get order for verification update",
			"order_id", payload.OrderID,
			"err", err,
		)
		return
	}

	if order.TopicID == nil || order.ThreadID == nil {
		logger.Log.Warnw("order has no topic or thread for control panel update",
			"order_id", order.ID,
		)
		return
	}

	messageID, ok := h.findControlPanelMessageID(
		ctx,
		order.ID,
		*order.TopicID,
	)
	if !ok {
		logger.Log.Warnw("control panel message not found, sending new one",
			"order_id", order.ID,
			"topic_id", *order.TopicID,
		)
		h.renderControlPanel(ctx, *order.TopicID, *order.ThreadID, order)
		return
	}

	h.renderEditControlPanel(
		ctx,
		messageID,
		*order.TopicID,
		*order.ThreadID,
		order,
	)
}
