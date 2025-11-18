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

func (s *OrderService) CreateOrder(ctx context.Context, userID int) (int, error) {
	return s.repo.Create(ctx, domain.User{ID: userID}, domain.Order{Status: "new"})
}

func (s *OrderService) Accept(
	ctx context.Context,
	assessorID int,
	orderID int,
	topicID, threadID int64,
) error {
	if err := s.repo.Accept(
		ctx,
		domain.Assessor{
			ID: assessorID,
		},
		orderID,
		"active",
		topicID,
		threadID,
	); err != nil {
		return fmt.Errorf("failed to accept order: %w", err)
	}
	return nil
}

func (s *OrderService) GetActiveByClient(ctx context.Context, userID int, status string) (*domain.Order, error) {
	order, err := s.repo.GetByUser(ctx, domain.User{ID: userID}, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return order, nil
}

func (s *OrderService) GetActiveByThread(ctx context.Context, topicID, threadID int64) (*domain.Order, error) {
	order, err := s.repo.GetByThread(ctx, topicID, threadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return order, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, id int) *domain.Order {
	order, _ := s.repo.GetByID(ctx, id)
	return order
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int, status string) {
	s.repo.UpdateStatus(ctx, id, status)
}
