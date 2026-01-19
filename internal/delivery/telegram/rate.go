package telegram

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleRateSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))
	parts := strings.Split(cb.Data, ":")
	if len(parts) < 3 {
		return
	}
	rate, _ := strconv.Atoi(parts[1])
	orderID, _ := strconv.Atoi(parts[2])

	h.reviewService.Rate(ctx, orderID, rate)

	h.stateService.SetStateWritingReview(ctx, chatID)

	h.bot.Send(tgbotapi.NewEditMessageText(
		chatID,
		messageID,
		h.text.WriteReviewText,
	))
}
