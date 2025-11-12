package usecase

import (
	"context"
	"fmt"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type OrderService struct {
	repo domain.OrderRepository
}

func NewOrderService(repo domain.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID int64) (int, error) {
	order := domain.Order{
		UserID: userID,
		Status: "new",
	}
	return s.repo.Create(ctx, order)
}

func (s *OrderService) Accept(
	ctx context.Context,
	appraiserID int64,
	orderID int,
	topicID, threadID int64,
) error {
	if err := s.repo.Accept(
		ctx,
		appraiserID,
		orderID,
		"active",
		topicID,
		threadID,
	); err != nil {
		return fmt.Errorf("failed to accept order: %w", err)
	}
	return nil
}

func (s *OrderService) GetActiveByClient(ctx context.Context, userID int64) (*domain.Order, error) {
	order, err := s.repo.GetByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return order, nil
}
