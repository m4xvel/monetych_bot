package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleTypeSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	logger.Log.Infow("game type selected",
		"chat_id", chatID,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 3 {
		logger.Log.Warnw("invalid type callback data",
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

	gameTypeID, err := strconv.Atoi(parts[2])
	if err != nil {
		logger.Log.Warnw("failed to parse game type id",
			"chat_id", chatID,
			"value", parts[2],
		)
		return
	}

	t, err := h.gameService.GetTypeByID(gameTypeID)
	if err != nil {
		logger.Log.Warnw("game type not found",
			"chat_id", chatID,
			"game_type_id", gameTypeID,
		)
		return
	}

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
	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn),
	)
	h.bot.Send(message)
}
