package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleDeclinedSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	messageID := cb.Message.MessageID

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))
	parts := strings.Split(cb.Data, ":")
	if len(parts) < 4 {
		return
	}
	orderID, _ := strconv.Atoi(parts[1])
	topicID, _ := strconv.ParseInt(parts[2], 10, 64)
	threadID, _ := strconv.ParseInt(parts[3], 10, 64)

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
		h.text.ConfirmDeclineText)

	editMessage.ReplyMarkup = &markup

	h.bot.Send(editMessage)
}
