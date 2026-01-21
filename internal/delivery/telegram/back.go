package telegram

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleBack(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	logger.Log.Infow("back navigation action initiated",
		"chat_id", chatID,
		"callback_data", cb.Data,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 4 {
		logger.Log.Warnw("invalid back callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	orderID, err := strconv.Atoi(parts[1])
	if err != nil {
		logger.Log.Warnw("failed to parse order_id from back callback",
			"chat_id", chatID,
			"value", parts[1],
		)
		return
	}

	topicID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		logger.Log.Warnw("failed to parse topic_id from back callback",
			"chat_id", chatID,
			"value", parts[2],
		)
		return
	}

	threadID, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		logger.Log.Warnw("failed to parse thread_id from back callback",
			"chat_id", chatID,
			"value", parts[3],
		)
		return
	}

	order, err := h.orderService.GetOrderByID(ctx, orderID)
	if err != nil {
		logger.Log.Errorw("failed to get order for back navigation",
			"chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
		return
	}

	h.renderEditControlPanel(
		cb.Message.MessageID,
		topicID,
		threadID,
		order,
	)

	logger.Log.Infow("back navigation ui rendered",
		"chat_id", chatID,
		"order_id", orderID,
	)
}
