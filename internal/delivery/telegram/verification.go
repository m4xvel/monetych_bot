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
	h.answerCallback(cb, h.text.VerificationRequestSentToast)

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
		if isInvalidToken(err) {
			logger.Log.Warnw("invalid verification callback token",
				"chat_id", chatID,
				"data", cb.Data,
				"err", err,
			)
			return
		}
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
		wrapped := wrapTelegramErr("telegram.send_verification_message", err)
		logger.Log.Errorw("failed to send verification message to user",
			"chat_id", payload.UserChatID,
			"order_id", payload.OrderID,
			"err", wrapped,
		)
		return
	}

	h.removeVerificationButton(cb.Message)
}

func (h *Handler) handleVerifySelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	h.answerCallback(cb, h.text.VerificationRequestReceivedToast)

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
		if isInvalidToken(err) {
			logger.Log.Warnw("invalid verify callback token",
				"chat_id", chatID,
				"data", cb.Data,
				"err", err,
			)
			return
		}
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

	if err := h.callbackTokenService.DeleteByActionAndOrderID(
		ctx,
		"verification",
		payload.OrderID,
	); err != nil {
		logger.Log.Errorw("failed to delete verification callbacks",
			"order_id", payload.OrderID,
			"err", err,
		)
	}

	edit := tgbotapi.NewEditMessageText(
		chatID,
		cb.Message.MessageID,
		h.text.SuccessfulVerify,
	)
	edit.ReplyMarkup = nil

	if _, err := h.bot.Send(edit); err != nil {
		wrapped := wrapTelegramErr("telegram.edit_verify_success", err)
		logger.Log.Errorw("failed to edit verification message",
			"chat_id", chatID,
			"err", wrapped,
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

func (h *Handler) removeVerificationButton(
	msg *tgbotapi.Message,
) {
	if msg == nil || msg.ReplyMarkup == nil {
		return
	}

	markup, changed := stripVerificationButton(msg.ReplyMarkup)
	if !changed {
		return
	}

	edit := tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      msg.Chat.ID,
			MessageID:   msg.MessageID,
			ReplyMarkup: markup,
		},
	}

	if _, err := h.bot.Send(edit); err != nil {
		wrapped := wrapTelegramErr("telegram.remove_verification_button", err)
		logger.Log.Errorw("failed to remove verification button",
			"chat_id", msg.Chat.ID,
			"message_id", msg.MessageID,
			"err", wrapped,
		)
	}
}

func stripVerificationButton(
	markup *tgbotapi.InlineKeyboardMarkup,
) (*tgbotapi.InlineKeyboardMarkup, bool) {
	if markup == nil {
		return nil, false
	}

	changed := false
	filtered := make([][]tgbotapi.InlineKeyboardButton, 0, len(markup.InlineKeyboard))

	for _, row := range markup.InlineKeyboard {
		newRow := make([]tgbotapi.InlineKeyboardButton, 0, len(row))
		for _, btn := range row {
			if btn.CallbackData != nil && strings.HasPrefix(*btn.CallbackData, "verification:") {
				changed = true
				continue
			}
			newRow = append(newRow, btn)
		}
		if len(newRow) > 0 {
			filtered = append(filtered, newRow)
		}
	}

	if !changed {
		return nil, false
	}

	if len(filtered) == 0 {
		return nil, true
	}

	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: filtered}, true
}
