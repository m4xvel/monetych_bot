package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

func (h *Handler) handleStartCommand(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: h.text.StartMenuText},
		{Command: "catalog", Description: h.text.CatalogMenuText},
		{Command: "support", Description: h.text.SupportMenuText},
		{Command: "reviews", Description: h.text.ReviewsMenuText},
	}

	h.bot.Request(tgbotapi.NewSetMyCommands(commands...))

	h.bot.SetChatMenuButton(tgbotapi.SetChatMenuButtonConfig{
		ChatID: chatID,
		MenuButton: tgbotapi.MenuButton{
			Type: "commands",
		},
	})

	h.userService.AddIfNotExists(ctx, chatID)

	h.bot.Send(tgbotapi.NewMessage(
		chatID,
		h.text.HelloText,
	))

	user, _ := h.userService.GetByTgID(ctx, chatID)
	h.stateService.SetState(ctx, domain.UserState{
		UserID: user.ID,
		State:  domain.StateStart,
	})
}
