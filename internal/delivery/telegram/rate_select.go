package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

func (h *Handler) handleRateSelect(
	ctx context.Context, cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	if !h.shouldProcess(chatID, messageID) {
		return
	}

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 3 {
		return
	}
	rate, _ := strconv.Atoi(parts[1])
	orderID, _ := strconv.Atoi(parts[2])

	user, _ := h.userService.GetByTgID(ctx, chatID)

	reviewID, err := h.reviewService.AddReview(ctx, orderID, user.ID, rate, "")
	if err != nil {
		fmt.Printf("insert review: %s", err)
	}

	h.stateService.SetState(ctx, domain.UserState{
		UserID:   user.ID,
		State:    domain.StateWritingReview,
		ReviewID: &reviewID,
	})

	h.bot.Send(tgbotapi.NewEditMessageText(
		chatID,
		messageID,
		h.text.WriteReviewText,
	))
}
