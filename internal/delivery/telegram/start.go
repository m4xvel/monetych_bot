package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleStartCommand(
	ctx context.Context,
	msg *tgbotapi.Message,
) {
	chat := msg.Chat
	chatID := chat.ID
	name := chat.FirstName

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

	h.userService.AddUser(
		ctx, chatID, name, func() string {
			return h.feature.GetUserAvatar(h.bot, chatID)
		},
	)

	h.bot.Send(tgbotapi.NewMessage(
		chatID,
		h.text.HelloText,
	))

	h.stateService.SetStateStart(ctx, chatID)
}
