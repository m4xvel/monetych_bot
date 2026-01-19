package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleGameSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 2 {
		return
	}

	gameID, _ := strconv.Atoi(parts[1])
	game, _ := h.gameService.GetGameByID(gameID)

	editText := tgbotapi.NewEditMessageText(
		chatID, cb.Message.MessageID, h.textDynamic.YouHaveChosenGame(game.Name),
	)
	h.bot.Request(editText)

	types, _ := h.gameService.GetGameTypesByGameID(gameID)
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
