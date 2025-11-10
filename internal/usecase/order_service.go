package usecase

import (
	"context"
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

type OrderService struct {
	repo domain.OrderRepository
	bot  *tgbotapi.BotAPI
}

func NewOrderService(repo domain.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID int64) (int, error) {
	order := domain.Order{
		UserID: userID,
		Status: "pending",
	}
	return s.repo.Create(ctx, order)
}

func (s *OrderService) Accept(ctx context.Context, appraiserID int64, orderID int) error {
	if err := s.repo.Accept(ctx, appraiserID, orderID, "active"); err != nil {
		return fmt.Errorf("failed to accept order: %w", err)
	}
	return nil
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
