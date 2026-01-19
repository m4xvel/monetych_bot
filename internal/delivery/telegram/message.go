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
	if msg.From != nil && msg.From.IsBot {
		return
	}

	if msg.MessageThreadID != 0 {
		h.handleExpertMessage(ctx, msg)
		return
	}

	h.handleUserMessage(ctx, msg)
}

func (h *Handler) handleUserMessage(
	ctx context.Context,
	msg *tgbotapi.Message,
) {
	chatID := msg.Chat.ID

	state, err := h.stateService.GetStateByChatID(ctx, chatID)
	if err != nil || state == nil {
		return
	}

	switch state.State {

	case domain.StateCommunication:
		h.forwardToExpert(state, msg)

	case domain.StateStart:
		h.handlerCatalogCommand(ctx, msg)

	case domain.StateWritingReview:
		if msg.Text == "" {
			return
		}

		h.reviewService.AddText(ctx, *state.ReviewID, msg.Text)
		h.bot.Send(tgbotapi.NewMessage(chatID, h.text.ThanksForReviewText))
		h.stateService.SetStateIdle(ctx, chatID)
	}
}

func (h *Handler) forwardToExpert(
	state *domain.UserState,
	msg *tgbotapi.Message,
) {
	if state.ExpertTopicID == nil || state.OrderThreadID == nil {
		return
	}

	params := tgbotapi.Params{
		"chat_id":           int64PtrToStr(state.ExpertTopicID),
		"from_chat_id":      fmt.Sprint(msg.Chat.ID),
		"message_id":        fmt.Sprint(msg.MessageID),
		"message_thread_id": int64PtrToStr(state.OrderThreadID),
	}

	h.bot.MakeRequest("copyMessage", params)
}

func (h *Handler) handleExpertMessage(
	ctx context.Context,
	msg *tgbotapi.Message,
) {
	isSystemMessage(msg)

	state, err := h.stateService.GetStateByThreadID(ctx, msg.MessageThreadID)
	if err != nil || state == nil {
		return
	}

	params := tgbotapi.Params{
		"chat_id":      int64PtrToStr(state.UserChatID),
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

func isSystemMessage(msg *tgbotapi.Message) bool {
	return msg.NewChatMembers != nil ||
		msg.LeftChatMember != nil ||
		msg.PinnedMessage != nil ||
		msg.NewChatTitle != "" ||
		msg.DeleteChatPhoto ||
		msg.GroupChatCreated ||
		msg.SuperGroupChatCreated ||
		msg.ChannelChatCreated
}
