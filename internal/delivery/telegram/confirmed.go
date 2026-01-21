package telegram

import (
	"context"
	"fmt"
	"strconv"
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
	if len(parts) < 4 {
		logger.Log.Warnw("invalid confirm callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	orderID, err := strconv.Atoi(parts[1])
	if err != nil {
		logger.Log.Warnw("failed to parse order_id from confirm callback",
			"chat_id", chatID,
			"value", parts[1],
		)
		return
	}

	topicID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		logger.Log.Warnw("failed to parse topic_id from confirm callback",
			"chat_id", chatID,
			"value", parts[2],
		)
		return
	}

	threadID, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		logger.Log.Warnw("failed to parse thread_id from confirm callback",
			"chat_id", chatID,
			"value", parts[3],
		)
		return
	}

	btn := tgbotapi.NewInlineKeyboardButtonData(
		h.text.AcceptText,
		fmt.Sprintf("confirmed_reaffirm:%d:%d:%d", orderID, topicID, threadID),
	)

	btnBack := tgbotapi.NewInlineKeyboardButtonData(
		"⬅️ Вернуться назад",
		fmt.Sprintf("back:%d:%d:%d", orderID, topicID, threadID),
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

	logger.Log.Infow("confirm order ui rendered",
		"chat_id", chatID,
		"order_id", orderID,
	)
}
