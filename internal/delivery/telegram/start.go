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

	experts, _ := h.expertService.GetAllExperts()
	for _, e := range experts {
		if chatID == e.ChatID {
			return
		}
	}

	support := h.supportService.GetSupport()
	if chatID == support.ChatID {

		commands := []tgbotapi.BotCommand{
			{Command: "search", Description: h.text.SupportMenuText},
		}

		scope := tgbotapi.NewBotCommandScopeChat(support.ChatID)
		cfg := tgbotapi.NewSetMyCommandsWithScope(scope, commands...)

		h.bot.Request(cfg)

		return
	}

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: h.text.StartMenuText},
		{Command: "catalog", Description: h.text.CatalogMenuText},
		{Command: "support", Description: h.text.SupportMenuText},
	}

	scope := tgbotapi.NewBotCommandScopeChat(chatID)
	cfg := tgbotapi.NewSetMyCommandsWithScope(scope, commands...)
	h.bot.Request(cfg)

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
