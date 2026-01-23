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

	support := h.supportService.GetSupport()

	message := tgbotapi.NewMessage(
		chatID,
		fmt.Sprintf("Служба технической поддержи: %s", support.ChatLink),
	)

	if _, err := h.bot.Send(message); err != nil {
		logger.Log.Errorw("failed to send support message",
			"chat_id", chatID,
			"err", err,
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
