package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleStartCommand(
	ctx context.Context,
	msg *tgbotapi.Message,
) {
	chatID := msg.Chat.ID

	logger.Log.Infow("start command initiated",
		"chat_id", chatID,
	)

	experts, err := h.expertService.GetAllExperts()
	if err != nil {
		logger.Log.Errorw("failed to get experts on start",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	for _, e := range experts {
		if chatID == e.ChatID {
			logger.Log.Infow("start command ignored for expert",
				"chat_id", chatID,
			)
			return
		}
	}

	support := h.supportService.GetSupport()
	if chatID == support.ChatID {
		logger.Log.Infow("start command for support",
			"chat_id", chatID,
		)

		commands := []tgbotapi.BotCommand{
			{Command: "search", Description: h.text.SupportMenuText},
		}

		scope := tgbotapi.NewBotCommandScopeChat(chatID)
		cfg := tgbotapi.NewSetMyCommandsWithScope(scope, commands...)

		if _, err := h.bot.Request(cfg); err != nil {
			logger.Log.Errorw("failed to set support commands",
				"chat_id", chatID,
				"err", err,
			)
		}

		return
	}

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: h.text.StartMenuText},
		{Command: "catalog", Description: h.text.CatalogMenuText},
		{Command: "support", Description: h.text.SupportMenuText},
	}

	scope := tgbotapi.NewBotCommandScopeChat(chatID)
	cfg := tgbotapi.NewSetMyCommandsWithScope(scope, commands...)

	if _, err := h.bot.Request(cfg); err != nil {
		logger.Log.Errorw("failed to set user commands",
			"chat_id", chatID,
			"err", err,
		)
	}

	h.bot.SetChatMenuButton(tgbotapi.SetChatMenuButtonConfig{
		ChatID: chatID,
		MenuButton: tgbotapi.MenuButton{
			Type: "commands",
		},
	})

	if err := h.userService.AddUser(
		ctx,
		chatID,
		msg.Chat.FirstName,
		func() string {
			return h.feature.GetUserAvatar(h.bot, chatID)
		},
	); err != nil {
		logger.Log.Errorw("failed to add user on start",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	logger.Log.Infow("user initialized on start",
		"chat_id", chatID,
	)

	if _, err := h.bot.Send(
		tgbotapi.NewMessage(chatID, h.text.HelloText),
	); err != nil {
		logger.Log.Errorw("failed to send hello message",
			"chat_id", chatID,
			"err", err,
		)
	}

	if err := h.stateService.SetStateStart(ctx, chatID); err != nil {
		logger.Log.Errorw("failed to set start state",
			"chat_id", chatID,
			"err", err,
		)
	}
}
