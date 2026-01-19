package usecase

import (
	"context"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type OrderMessageService struct {
	repo domain.OrderMessageRepository
}

func NewOrderMessageService(
	r domain.OrderMessageRepository,
) *OrderMessageService {
	return &OrderMessageService{repo: r}
}

func (s *OrderMessageService) Save(
	ctx context.Context,
	orderID int,
	chatID int64,
	messageID int,
) error {
	return s.repo.Save(ctx, domain.OrderMessage{
		OrderID:   orderID,
		ChatID:    chatID,
		MessageID: messageID,
	})
}

func (s *OrderMessageService) GetByOrder(
	ctx context.Context,
	orderID int,
) ([]domain.OrderMessage, error) {
	return s.repo.Get(ctx, orderID)
}

func (s *OrderMessageService) MarkDeletedByOrder(
	ctx context.Context,
	orderID int,
) error {
	return s.repo.Delete(ctx, orderID)
}
