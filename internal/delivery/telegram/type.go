package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleTypeSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 3 {
		return
	}

	gameID, _ := strconv.Atoi(parts[1])
	gameTypeID, _ := strconv.Atoi(parts[2])

	t, _ := h.gameService.GetTypeByID(gameTypeID)

	editText := tgbotapi.NewEditMessageText(
		chatID,
		cb.Message.MessageID,
		h.textDynamic.YouHaveChosenType(t.Name),
	)
	h.bot.Request(editText)

	message := tgbotapi.NewMessage(chatID, h.text.ContactAppraiserText)
	btn := tgbotapi.NewInlineKeyboardButtonData(
		h.text.ContactText,
		fmt.Sprintf("order:%d:%d", gameID, gameTypeID),
	)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn),
	)
	message.ReplyMarkup = keyboard
	h.bot.Send(message)
}
