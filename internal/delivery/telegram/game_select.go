package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleGameSelect(
	ctx context.Context, cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	if !h.shouldProcess(chatID, messageID) {
		return
	}

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 3 {
		return
	}
	gameID, _ := strconv.Atoi(parts[1])
	gameName := parts[2]

	editText := tgbotapi.NewEditMessageText(
		chatID,
		cb.Message.MessageID,
		h.textDynamic.YouHaveChosenGame(gameName),
	)
	h.bot.Request(editText)

	types, _ := h.gameService.ListGameTypes(ctx, gameID)

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, t := range types {
		data := fmt.Sprintf("type:%s:%s", gameName, t)
		btn := tgbotapi.NewInlineKeyboardButtonData(t, data)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	msg := tgbotapi.NewMessage(chatID, h.text.ChooseType)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.bot.Send(msg)
}
