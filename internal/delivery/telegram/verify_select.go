package telegram

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleVerifySelect(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	if !h.shouldProcess(chatID, messageID) {
		return
	}

	_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 4 {
		return
	}

	isVerified, _ := strconv.ParseBool(parts[1])
	itemGame := parts[2]
	itemType := parts[3]

	if isVerified {
		h.userService.VerifyUser(ctx, chatID)
		editText := tgbotapi.NewEditMessageText(
			chatID,
			cb.Message.MessageID,
			"✅ Ваша личность подтверждена!",
		)
		_, _ = h.bot.Request(editText)

		h.contactAnAppraiser(ctx, chatID, itemGame, itemType)
		return
	}
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	h.bot.Send(deleteMsg)
	h.showInlineKeyboardVerification(
		chatID,
		"❌ Вы не прошли верификацию, попробуйте снова!",
		true,
		itemGame,
		itemType,
	)
}
