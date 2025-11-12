package telegram

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

func (h *Handler) handleMessage(
	ctx context.Context,
	msg *tgbotapi.Message,
) {
	if msg.MessageThreadID != 0 {
		h.handleAssessorMessage(ctx, msg)
		return
	}

	h.handleUserMessage(ctx, msg)
}

func (h *Handler) handleAssessorMessage(ctx context.Context, msg *tgbotapi.Message) {
	threadID := msg.MessageThreadID
	order, err := h.orderService.GetActiveByThread(ctx, msg.Chat.ID, threadID)
	if err != nil {
		log.Println("failed to get order by thread:", err)
		return
	}
	if order != nil && order.Status == "active" {
		h.forwardToUser(order, msg)
	}
}

func (h *Handler) handleUserMessage(ctx context.Context, msg *tgbotapi.Message) {
	order, err := h.orderService.GetActiveByClient(ctx, msg.From.ID)
	if err != nil {
		log.Println("failed to get order by client:", err)
		return
	}
	if order != nil && order.Status == "active" {
		h.forwardToAssessor(order, msg)
	}
}

func (h *Handler) forwardToAssessor(order *domain.Order, msg *tgbotapi.Message) {
	params := tgbotapi.Params{
		"chat_id":           fmt.Sprint(order.TopicID),
		"from_chat_id":      fmt.Sprint(msg.Chat.ID),
		"message_id":        fmt.Sprint(msg.MessageID),
		"message_thread_id": fmt.Sprint(order.ThreadID),
	}
	h.bot.MakeRequest("copyMessage", params)
}

func (h *Handler) forwardToUser(order *domain.Order, msg *tgbotapi.Message) {
	params := tgbotapi.Params{
		"chat_id":      fmt.Sprint(order.UserID),
		"from_chat_id": fmt.Sprint(msg.Chat.ID),
		"message_id":   fmt.Sprint(msg.MessageID),
	}
	h.bot.MakeRequest("copyMessage", params)
}
