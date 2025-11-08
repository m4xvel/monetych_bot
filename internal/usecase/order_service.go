package usecase

import (
	"context"
	"errors"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type OrderService struct {
	repo domain.OrderRepository
}

func NewOrderService(repo domain.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID, appraiserID int64) (int, error) {
	order := domain.Order{
		UserID:      userID,
		AppraiserID: appraiserID,
		Status:      "active",
	}
	return s.repo.Create(ctx, order)
}

func (s *OrderService) GetUserOrder(ctx context.Context, userID int64) (*domain.Order, error) {
	order, err := s.repo.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if order == nil || order.Status != "active" {
		return nil, errors.New("no active order found")
	}
	return order, nil
}
