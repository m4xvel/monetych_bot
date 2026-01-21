package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleGameSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	logger.Log.Infow("game selected",
		"chat_id", chatID,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 2 {
		logger.Log.Warnw("invalid game callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	gameID, err := strconv.Atoi(parts[1])
	if err != nil {
		logger.Log.Warnw("failed to parse game id",
			"chat_id", chatID,
			"value", parts[1],
		)
		return
	}

	game, err := h.gameService.GetGameByID(gameID)
	if err != nil {
		logger.Log.Warnw("game not found",
			"chat_id", chatID,
			"game_id", gameID,
		)
		return
	}

	editText := tgbotapi.NewEditMessageText(
		chatID, cb.Message.MessageID, h.textDynamic.YouHaveChosenGame(game.Name),
	)
	h.bot.Request(editText)

	types, err := h.gameService.GetGameTypesByGameID(gameID)
	if err != nil {
		logger.Log.Errorw("failed to get game types",
			"chat_id", chatID,
			"game_id", gameID,
			"err", err,
		)
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, t := range types {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			t.Name,
			fmt.Sprintf("type:%d:%d", game.ID, t.ID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	message := tgbotapi.NewMessage(chatID, h.text.ChooseType)
	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.bot.Send(message)
}
