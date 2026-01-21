package telegram

import (
	"context"
	"fmt"
	"strconv"
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
	if len(parts) < 4 {
		logger.Log.Warnw("invalid declined callback data",
			"data", cb.Data,
		)
		return
	}

	orderID, err := strconv.Atoi(parts[1])
	if err != nil {
		logger.Log.Warnw("failed to parse order id for decline",
			"value", parts[1],
		)
		return
	}

	topicID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		logger.Log.Warnw("failed to parse topic id for decline",
			"value", parts[2],
		)
		return
	}

	threadID, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		logger.Log.Warnw("failed to parse thread id for decline",
			"value", parts[3],
		)
		return
	}

	logger.Log.Infow("decline confirmation requested",
		"order_id", orderID,
		"topic_id", topicID,
		"thread_id", threadID,
	)

	btn := tgbotapi.NewInlineKeyboardButtonData(
		h.text.AcceptText,
		fmt.Sprintf("declined_reaffirm:%d:%d:%d", orderID, topicID, threadID),
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
		h.text.ConfirmDeclineText,
	)
	editMessage.ReplyMarkup = &markup

	h.bot.Send(editMessage)
}
