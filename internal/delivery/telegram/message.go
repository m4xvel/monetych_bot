package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

func (h *Handler) handleMessage(
	ctx context.Context,
	msg *tgbotapi.Message,
) {
	if msg.MessageThreadID != 0 {
		h.handleAssessorMessage(ctx, msg)
		return
	}
	h.handleUserMessage(ctx, msg)
}

func (h *Handler) handleAssessorMessage(ctx context.Context, msg *tgbotapi.Message) {
	threadID := msg.MessageThreadID
	order, _ := h.orderService.GetActiveByThread(ctx, msg.Chat.ID, threadID)
	if order == nil {
		return
	}
	user, _ := h.userService.GetByID(ctx, order.UserID)
	if order.Status == domain.OrderActive {
		h.forwardToUser(user, msg)
	}
}

func (h *Handler) handleUserMessage(ctx context.Context, msg *tgbotapi.Message) {
	user, _ := h.userService.GetByTgID(ctx, msg.From.ID)
	order, _ := h.orderService.GetActiveByClient(ctx, user.ID)

	if order != nil && order.Status == domain.OrderActive {
		h.forwardToAssessor(order, msg)
		return
	}

	state, _ := h.stateService.GetState(ctx, user.ID)
	fmt.Println(state)
	if state == nil {
		return
	}

	if state.State == domain.StateStart {
		h.handleCatalogCommand(ctx, msg)
		return
	}

	if state.State == domain.StateWritingReview {
		h.reviewService.UpdateText(ctx, *state.ReviewID, msg.Text)
		h.bot.Send(tgbotapi.NewMessage(user.UserID, h.text.ThanksForReviewText))
		h.stateService.SetState(ctx, domain.UserState{
			UserID: user.ID,
			State:  domain.StateIdle,
		})
		return
	}
}

func (h *Handler) forwardToAssessor(order *domain.Order, msg *tgbotapi.Message) {
	params := tgbotapi.Params{
		"chat_id":           int64PtrToStr(order.TopicID),
		"from_chat_id":      fmt.Sprint(msg.Chat.ID),
		"message_id":        fmt.Sprint(msg.MessageID),
		"message_thread_id": int64PtrToStr(order.ThreadID),
	}
	h.bot.MakeRequest("copyMessage", params)
}

func (h *Handler) forwardToUser(user *domain.User, msg *tgbotapi.Message) {
	params := tgbotapi.Params{
		"chat_id":      fmt.Sprint(user.UserID),
		"from_chat_id": fmt.Sprint(msg.Chat.ID),
		"message_id":   fmt.Sprint(msg.MessageID),
	}
	h.bot.MakeRequest("copyMessage", params)
}

func int64PtrToStr(v *int64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(*v)
}
