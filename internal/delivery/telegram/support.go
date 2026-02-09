package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handlerSupportCommand(
	ctx context.Context,
	msg *tgbotapi.Message,
) {
	chatID := msg.Chat.ID

	supportInfo := h.supportService.GetSupport()

	message := tgbotapi.NewMessage(
		chatID,
		fmt.Sprintf(h.text.SupportContactTemplate, supportInfo.ChatLink),
	)

	if _, err := h.bot.Send(message); err != nil {
		wrapped := wrapTelegramErr("telegram.send_support_message", err)
		logger.Log.Errorw("failed to send support message",
			"chat_id", chatID,
			"err", wrapped,
		)
		return
	}

	if err := h.stateService.SetStateIdle(ctx, chatID); err != nil {
		logger.Log.Errorw("failed to set idle state after support command",
			"chat_id", chatID,
			"err", err,
		)
		return
	}
}
