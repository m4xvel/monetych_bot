package telegram

import (
	"context"
	"fmt"

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
		btn := tgbotapi.NewInlineKeyboardButtonData(
			g.Name,
			fmt.Sprintf("game:%d", g.ID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	message := tgbotapi.NewMessage(chatID, h.text.ChooseGame)
	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	if _, err := h.bot.Send(message); err != nil {
		logger.Log.Errorw("failed to send catalog message",
			"chat_id", chatID,
			"err", err,
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

	logger.Log.Infow("catalog ui rendered",
		"chat_id", chatID,
		"games_count", len(games),
	)
}
