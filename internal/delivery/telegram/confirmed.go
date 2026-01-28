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
	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

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

	h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"confirmed",
		&payload,
	)

	orderID := payload.OrderID
	topicID := payload.TopicID
	threadID := payload.ThreadID

	h.callbackTokenService.DeleteByActionAndOrderID(
		ctx,
		"declined",
		orderID,
	)

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
		h.text.ConfirmConfirmedText,
	)
	editMessage.ReplyMarkup = &markup

	if _, err := h.bot.Send(editMessage); err != nil {
		logger.Log.Errorw("failed to edit confirmation message",
			"chat_id", chatID,
			"order_id", orderID,
			"topic_id", topicID,
			"err", err,
		)
		return
	}
}
