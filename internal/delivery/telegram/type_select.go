package telegram

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleTypeSelect(ctx context.Context, cb *tgbotapi.CallbackQuery) {
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
	itemGame := parts[1]
	itemType := parts[2]

	editText := tgbotapi.NewEditMessageText(
		chatID,
		cb.Message.MessageID,
		h.textDynamic.YouHaveChosenType(itemType),
	)
	h.bot.Request(editText)

	isVerified := h.userService.CheckStatusVerified(ctx, chatID)
	if !isVerified {
		h.showInlineKeyboardVerification(
			chatID,
			h.text.YouNeedToVerify,
			false,
			itemGame,
			itemType,
		)
		return
	}

	h.contactAnAppraiser(ctx, chatID, itemGame, itemType)
}
