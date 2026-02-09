package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handlerCatalogCommand(
	ctx context.Context,
	msg *tgbotapi.Message,
) {
	chatID := msg.Chat.ID

	logger.Log.Infow("catalog command initiated",
		"chat_id", chatID,
	)

	games, err := h.gameService.GetAllGames()
	if err != nil {
		logger.Log.Errorw("failed to get games for catalog",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, g := range games {

		token, err := h.callbackTokenService.Create(
			ctx,
			"game",
			&GameSelectPayload{
				GameID: g.ID,
				ChatID: chatID,
			},
		)
		if err != nil {
			logger.Log.Errorw("failed to create game callback token",
				"chat_id", chatID,
				"err", err,
			)
			continue
		}

		btn := tgbotapi.NewInlineKeyboardButtonData(
			g.Name,
			"game:"+token,
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	message := tgbotapi.NewMessage(chatID, h.text.ChooseGame)
	message.ParseMode = "markdown"
	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	if _, err := h.bot.Send(message); err != nil {
		wrapped := wrapTelegramErr("telegram.send_catalog", err)
		logger.Log.Errorw("failed to send catalog message",
			"chat_id", chatID,
			"err", wrapped,
		)
		return
	}

	if err := h.stateService.SetStateIdle(ctx, chatID); err != nil {
		logger.Log.Errorw("failed to set idle state after catalog command",
			"chat_id", chatID,
			"err", err,
		)
		return
	}
}
