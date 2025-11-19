package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

func (h *Handler) handleSupportCommand(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	h.bot.Send(tgbotapi.NewMessage(
		chatID,
		h.text.SupportText,
	))

	user, _ := h.userService.GetByTgID(ctx, chatID)
	h.stateService.SetState(ctx, domain.UserState{
		UserID: user.ID,
		State:  domain.StateIdle,
	})
}
