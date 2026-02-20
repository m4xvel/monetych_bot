package telegram

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleGameSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	h.answerCallback(cb, "")

	logger.Log.Infow("game selected",
		"chat_id", chatID,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) != 2 {
		logger.Log.Warnw("invalid game callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	token := parts[1]

	var payload GameSelectPayload

	if err := h.callbackTokenService.Consume(
		ctx,
		token,
		"game",
		&payload,
	); err != nil {
		if isInvalidToken(err) {
			logger.Log.Warnw("invalid game callback token",
				"chat_id", chatID,
				"data", cb.Data,
				"err", err,
			)
			return
		}
		logger.Log.Errorw("failed to consume game callback token",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	if payload.ChatID != cb.From.ID {
		return
	}

	gameID := payload.GameID

	game, err := h.gameService.GetGameByID(gameID)
	if err != nil {
		logger.Log.Warnw("game not found",
			"chat_id", chatID,
			"game_id", gameID,
		)
		return
	}

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
	createdTokens := make([]string, 0, len(types))
	for _, t := range types {
		token, err := h.callbackTokenService.Create(
			ctx,
			"type",
			&TypeSelectPayload{
				ChatID: chatID,
				GameID: game.ID,
				TypeID: t.ID,
			},
		)
		if err != nil {
			logger.Log.Errorw("failed to create games type callback token",
				"chat_id", chatID,
				"err", err,
			)
			continue
		}
		createdTokens = append(createdTokens, token)

		btn := tgbotapi.NewInlineKeyboardButtonData(
			t.Name,
			"type:"+token,
		)

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	text := fmt.Sprintf(
		"%s\n\n%s",
		h.textDynamic.YouHaveChosenGame(game.Name),
		h.text.ChooseType,
	)

	edit := tgbotapi.NewEditMessageText(
		chatID,
		cb.Message.MessageID,
		text,
	)

	edit.ParseMode = "markdown"

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	edit.ReplyMarkup = &markup

	if _, err := h.bot.Request(edit); err != nil {
		wrapped := wrapTelegramErr("telegram.edit_game_message", err)
		logger.Log.Errorw("failed to edit game selection message",
			"chat_id", chatID,
			"err", wrapped,
		)
		for _, token := range createdTokens {
			if err := h.callbackTokenService.Delete(ctx, token, "type"); err != nil {
				logger.Log.Errorw("failed to cleanup type callback token",
					"chat_id", chatID,
					"err", err,
				)
			}
		}
	}
}
