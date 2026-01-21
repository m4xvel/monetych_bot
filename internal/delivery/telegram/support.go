package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	h.bot.Send(message)

	h.stateService.SetStateIdle(ctx, chatID)
}
