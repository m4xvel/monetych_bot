package telegram

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleConfirmedSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID
	h.answerCallback(cb, "")

	logger.Log.Infow("confirm order ui step initiated",
		"chat_id", chatID,
		"callback_data", cb.Data,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) != 2 {
		logger.Log.Warnw("invalid confirm callback data",
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
		"confirmed",
		&payload,
	); err != nil {
		if isInvalidToken(err) {
			logger.Log.Warnw("invalid confirmed callback token",
				"chat_id", chatID,
				"data", cb.Data,
				"err", err,
			)
			return
		}
		logger.Log.Errorw("failed to consume confirmed callback token",
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
		"declined",
		orderID,
	); err != nil {
		logger.Log.Errorw("failed to delete declined callbacks",
			"order_id", orderID,
			"err", err,
		)
	}

	tokenReaffirm, err := h.callbackTokenService.Create(
		ctx,
		"confirmed_reaffirm",
		&ConfirmedAndDeclinedOrderSelectPayload{
			OrderID:  orderID,
			TopicID:  topicID,
			ThreadID: threadID,
		},
	)
	if err != nil {
		logger.Log.Errorw(
			"failed to create confirmed reaffirm order callback token",
			"err", err,
		)
	}

	btn := tgbotapi.NewInlineKeyboardButtonData(
		h.text.AcceptText,
		"confirmed_reaffirm:"+tokenReaffirm,
	)

	tokenBack, err := h.callbackTokenService.Create(
		ctx,
		"back",
		&ConfirmedAndDeclinedOrderSelectPayload{
			OrderID:  orderID,
			TopicID:  topicID,
			ThreadID: threadID,
		},
	)
	if err != nil {
		logger.Log.Errorw(
			"failed to create back order callback token",
			"err", err,
		)
	}

	btnBack := tgbotapi.NewInlineKeyboardButtonData(
		h.text.BackButtonText,
		"back:"+tokenBack,
	)

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn),
		tgbotapi.NewInlineKeyboardRow(btnBack),
	)

	editMessage := tgbotapi.NewEditMessageText(
		topicID,
		messageID,
		h.text.ConfirmConfirmedText,
	)
	editMessage.ReplyMarkup = &markup

	if _, err := h.bot.Send(editMessage); err != nil {
		wrapped := wrapTelegramErr("telegram.edit_confirm_confirmation", err)
		logger.Log.Errorw("failed to edit confirmation message",
			"chat_id", chatID,
			"order_id", orderID,
			"topic_id", topicID,
			"err", wrapped,
		)
		if err := h.callbackTokenService.Delete(
			ctx,
			tokenReaffirm,
			"confirmed_reaffirm",
		); err != nil {
			logger.Log.Errorw("failed to cleanup confirmed reaffirm callback token",
				"order_id", orderID,
				"err", err,
			)
		}
		if err := h.callbackTokenService.Delete(ctx, tokenBack, "back"); err != nil {
			logger.Log.Errorw("failed to cleanup back callback token",
				"order_id", orderID,
				"err", err,
			)
		}
		return
	}
}
