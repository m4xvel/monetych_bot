package telegram

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleAcceptPrivacySelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	h.answerCallback(cb, "")

	parts := strings.Split(cb.Data, ":")
	if len(parts) != 2 {
		logger.Log.Warnw("invalid accept client callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	tokenCallback := parts[1]

	var payload AcceptPrivacySelectPayload

	if err := h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"accept_privacy",
		&payload,
	); err != nil {
		if isInvalidToken(err) {
			logger.Log.Warnw("invalid accept privacy callback token",
				"chat_id", chatID,
				"data", cb.Data,
				"err", err,
			)
			return
		}
		logger.Log.Errorw("failed to consume accept privacy callback token",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	userID := payload.ChatID

	if err := h.userService.AddUser(
		ctx,
		userID,
		cb.From.FirstName,
		func() string {
			return h.feature.GetUserAvatar(h.bot, chatID)
		},
	); err != nil {
		logger.Log.Errorw("failed to add user on accept privacy",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	logger.Log.Infow("user initialized on start",
		"chat_id", chatID,
	)

	if err := h.userPolicyAcceptancesService.Accept(
		ctx,
		userID,
	); err != nil {
		logger.Log.Errorw("failed to accept privacy on accept privacy",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	edit := tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      chatID,
			MessageID:   cb.Message.MessageID,
			ReplyMarkup: nil,
		},
	}

	_, err := h.bot.Send(edit)
	if err != nil {
		wrapped := wrapTelegramErr("telegram.remove_accept_privacy_keyboard", err)
		logger.Log.Errorw("failed to remove keyboard", "err", wrapped)
	}

	h.handlerCatalogCommand(ctx, cb.Message)
}
