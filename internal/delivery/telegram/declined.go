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
	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

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

	h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"declined",
		&payload,
	)

	orderID := payload.OrderID
	topicID := payload.TopicID
	threadID := payload.ThreadID

	h.callbackTokenService.DeleteByActionAndOrderID(
		ctx,
		"confirmed",
		orderID,
	)

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
		"⬅️ Вернуться назад",
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

	h.bot.Send(editMessage)
}
