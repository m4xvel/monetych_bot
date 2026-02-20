package telegram

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleDeclinedSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	messageID := cb.Message.MessageID
	h.answerCallback(cb, "")

	logger.Log.Infow("decline order initiated",
		"callback_data", cb.Data,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) != 2 {
		logger.Log.Warnw("invalid declined callback data",
			"data", cb.Data,
		)
		return
	}

	tokenCallback := parts[1]

	var payload ConfirmedAndDeclinedOrderSelectPayload

	if err := h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"declined",
		&payload,
	); err != nil {
		if isInvalidToken(err) {
			logger.Log.Warnw("invalid declined callback token",
				"data", cb.Data,
				"err", err,
			)
			return
		}
		logger.Log.Errorw("failed to consume declined callback token",
			"err", err,
		)
		return
	}

	orderID := payload.OrderID
	topicID := payload.TopicID
	threadID := payload.ThreadID

	if err := h.callbackTokenService.DeleteByActionAndOrderID(
		ctx,
		"confirmed",
		orderID,
	); err != nil {
		logger.Log.Errorw("failed to delete confirmed callbacks",
			"order_id", orderID,
			"err", err,
		)
	}

	logger.Log.Infow("decline confirmation requested",
		"order_id", orderID,
		"topic_id", topicID,
		"thread_id", threadID,
	)

	tokenReaffirm, err := h.callbackTokenService.Create(
		ctx,
		"declined_reaffirm",
		&ConfirmedAndDeclinedOrderSelectPayload{
			OrderID:  orderID,
			TopicID:  topicID,
			ThreadID: threadID,
		},
	)
	if err != nil {
		logger.Log.Errorw(
			"failed to create declined reaffirm order callback token",
			"err", err,
		)
	}

	btn := tgbotapi.NewInlineKeyboardButtonData(
		h.text.AcceptText,
		"declined_reaffirm:"+tokenReaffirm,
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
		h.text.ConfirmDeclineText,
	)
	editMessage.ReplyMarkup = &markup

	if _, err := h.bot.Send(editMessage); err != nil {
		wrapped := wrapTelegramErr("telegram.edit_decline_confirmation", err)
		logger.Log.Errorw("failed to edit decline confirmation message",
			"order_id", orderID,
			"topic_id", topicID,
			"err", wrapped,
		)
		if err := h.callbackTokenService.Delete(
			ctx,
			tokenReaffirm,
			"declined_reaffirm",
		); err != nil {
			logger.Log.Errorw("failed to cleanup declined reaffirm callback token",
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
	}
}
