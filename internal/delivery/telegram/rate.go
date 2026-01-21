package telegram

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleRateSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID
	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	logger.Log.Infow("rate order action initiated",
		"chat_id", chatID,
		"callback_data", cb.Data,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 3 {
		logger.Log.Warnw("invalid rate callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	rate, err := strconv.Atoi(parts[1])
	if err != nil {
		logger.Log.Warnw("failed to parse rate value",
			"chat_id", chatID,
			"value", parts[1],
		)
		return
	}

	if rate < 1 || rate > 5 {
		logger.Log.Warnw("rate value out of allowed range",
			"chat_id", chatID,
			"rate", rate,
		)
		return
	}

	orderID, err := strconv.Atoi(parts[2])
	if err != nil {
		logger.Log.Warnw("failed to parse order_id from rate callback",
			"chat_id", chatID,
			"value", parts[2],
		)
		return
	}

	if err := h.reviewService.Rate(ctx, orderID, rate); err != nil {
		logger.Log.Errorw("failed to rate order",
			"chat_id", chatID,
			"order_id", orderID,
			"rate", rate,
			"err", err,
		)
		return
	}

	logger.Log.Infow("order rated successfully",
		"chat_id", chatID,
		"order_id", orderID,
		"rate", rate,
	)

	if err := h.stateService.SetStateWritingReview(ctx, chatID); err != nil {
		logger.Log.Errorw("failed to set state writing review",
			"chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
		return
	}

	if _, err := h.bot.Send(
		tgbotapi.NewEditMessageText(
			chatID,
			messageID,
			h.text.WriteReviewText,
		),
	); err != nil {
		logger.Log.Errorw("failed to prompt user to write review",
			"chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
	}
}
