package telegram

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleClientMessage(
	ctx context.Context,
	msg *tgbotapi.Message,
) {
	userID := msg.From.ID
	order, err := h.orderService.GetActiveByClient(ctx, userID)
	if err != nil {
		log.Println("failed to get order", err)
		return
	}
	if order == nil {
		return
	}

	if order.Status != "active" {
		return
	}

	params := tgbotapi.Params{
		"chat_id":           fmt.Sprint(order.TopicID),
		"from_chat_id":      fmt.Sprint(msg.Chat.ID),
		"message_id":        fmt.Sprint(msg.MessageID),
		"message_thread_id": fmt.Sprint(order.ThreadID),
	}

	h.bot.MakeRequest("copyMessage", params)
}
