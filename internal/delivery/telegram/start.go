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
		if chatID == e.TopicID {
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

	accepted, _ := h.userPolicyAcceptancesService.IsAccepted(ctx, chatID)

	message := tgbotapi.NewMessage(chatID, h.textDynamic.HelloText())
	message.ParseMode = "Markdown"
	message.DisableWebPagePreview = true
	if !accepted {
		token, err := h.callbackTokenService.Create(
			ctx,
			"accept_privacy",
			&AcceptPrivacySelectPayload{
				ChatID: chatID,
			},
		)
		if err != nil {
			logger.Log.Errorw("failed to create accept privacy callback token",
				"chat_id", chatID,
				"err", err,
			)
		}

		message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(
				"Соглашаюсь",
				"accept_privacy:"+token,
			)),
		)
	}

	if _, err := h.bot.Send(message); err != nil {
		logger.Log.Errorw("failed to send hello message",
			"chat_id", chatID,
			"err", err,
		)
	}

	if accepted {
		h.handlerCatalogCommand(ctx, msg)
	}
}
