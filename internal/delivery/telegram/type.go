package telegram

import (
	"context"
	"fmt"
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
	if len(parts) != 2 {
		logger.Log.Warnw("invalid type callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	tokenCallback := parts[1]

	var payload TypeSelectPayload

	h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"type",
		&payload,
	)

	if payload.ChatID != cb.From.ID {
		return
	}

	gameID := payload.GameID
	gameTypeID := payload.TypeID

	t, err := h.gameService.GetTypeByID(gameTypeID)
	if err != nil {
		logger.Log.Warnw("game type not found",
			"chat_id", chatID,
			"game_type_id", gameTypeID,
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

	token, err := h.callbackTokenService.Create(
		ctx,
		"order",
		&OrderSelectPayload{
			ChatID: chatID,
			GameID: gameID,
			TypeID: gameTypeID,
		},
	)
	if err != nil {
		logger.Log.Errorw("failed to create games type callback token",
			"chat_id", chatID,
			"err", err,
		)
	}

	text := fmt.Sprintf(
		"%s\n\n%s",
		h.textDynamic.YouHaveChosenGameAndType(game.Name, t.Name),
		h.text.ContactAppraiserText,
	)

	edit := tgbotapi.NewEditMessageText(
		chatID,
		cb.Message.MessageID,
		text,
	)

	edit.ParseMode = "markdown"

	btn := tgbotapi.NewInlineKeyboardButtonData(
		h.text.ContactText,
		"order:"+token,
	)

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn),
	)

	edit.ReplyMarkup = &markup

	h.bot.Request(edit)
}
