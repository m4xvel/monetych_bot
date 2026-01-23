package usecase

import (
	"context"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type OrderChatMessageService struct {
	repo domain.OrderChatMessagesRepository
}

func NewOrderChatMessageService(
	r domain.OrderChatMessagesRepository,
) *OrderChatMessageService {
	return &OrderChatMessageService{
		repo: r,
	}
}

func (s *OrderChatMessageService) SaveUserMessage(
	ctx context.Context,
	orderID int,
	userID int,
	chatID int64,
	messageID int,
	msgType domain.MessageType,
	text *string,
	media map[string]any,
	raw map[string]any,
) error {
	msg := &domain.OrderChatMessages{
		OrderID:      orderID,
		SenderRole:   domain.SenderUser,
		SenderUserID: &userID,
		ChatID:       chatID,
		MessageID:    messageID,
		MessageType:  msgType,
		Text:         text,
		Media:        media,
		RawPayload:   raw,
	}

	return s.repo.Save(ctx, msg)
}

func (s *OrderChatMessageService) SaveExpertMessage(
	ctx context.Context,
	orderID int,
	expertID int,
	chatID int64,
	messageID int,
	msgType domain.MessageType,
	text *string,
	media map[string]any,
	raw map[string]any,
) error {
	msg := &domain.OrderChatMessages{
		OrderID:        orderID,
		SenderRole:     domain.SenderExpert,
		SenderExpertID: &expertID,
		ChatID:         chatID,
		MessageID:      messageID,
		MessageType:    msgType,
		Text:           text,
		Media:          media,
		RawPayload:     raw,
	}

	return s.repo.Save(ctx, msg)
}
