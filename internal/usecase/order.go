package usecase

import (
	"context"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type OrderService struct {
	orderRepo domain.OrderRepository
	userRepo  domain.UserRepository
}

func NewOrderService(
	or domain.OrderRepository,
	ur domain.UserRepository,
) *OrderService {
	return &OrderService{
		orderRepo: or,
		userRepo:  ur,
	}
}

func (s *OrderService) CreateOrder(
	ctx context.Context,
	userID, gameID, gameTypeID int,
	userNameAtPurchase, gameNameAtPurchase,
	gameTypeNameAtPurchase string,
) (int, error) {
	return s.orderRepo.Create(ctx, domain.Order{
		UserID:                 userID,
		GameID:                 gameID,
		GameTypeID:             gameTypeID,
		UserNameAtPurchase:     userNameAtPurchase,
		GameNameAtPurchase:     gameNameAtPurchase,
		GameTypeNameAtPurchase: gameTypeNameAtPurchase,
	})
}

func (s *OrderService) SetExpertData(
	ctx context.Context,
	orderID int,
	expertID int,
	threadID int64,
) error {
	return s.orderRepo.SetActive(
		ctx, domain.Order{
			ID:       orderID,
			ExpertID: &expertID,
			ThreadID: &threadID,
		},
		domain.OrderAccepted,
	)
}

func (s *OrderService) SetAcceptedStatus(ctx context.Context, orderID int) error {
	return s.orderRepo.UpdateStatus(
		ctx, domain.Order{
			ID:     orderID,
			Status: domain.OrderAccepted,
		},
		domain.OrderNew,
	)
}

func (s *OrderService) SetExpertConfirmedStatus(ctx context.Context, orderID int) error {
	return s.orderRepo.UpdateStatus(
		ctx, domain.Order{
			ID:     orderID,
			Status: domain.OrderExpertConfirmed,
		},
		domain.OrderAccepted,
	)
}

func (s *OrderService) SetCompletedStatus(
	ctx context.Context,
	orderID int,
	chatID int64,
) error {
	s.userRepo.IncrementOrders(ctx, chatID)
	return s.orderRepo.UpdateStatus(
		ctx, domain.Order{
			ID:     orderID,
			Status: domain.OrderCompleted,
		},
		domain.OrderExpertConfirmed,
	)
}

func (s *OrderService) SetCancelStatus(ctx context.Context, orderID int) error {
	return s.orderRepo.UpdateStatus(
		ctx, domain.Order{
			ID:     orderID,
			Status: domain.OrderCanceled,
		},
		domain.OrderNew,
	)
}

func (s *OrderService) SetDeclinedStatus(ctx context.Context, orderID int) error {
	return s.orderRepo.UpdateStatus(
		ctx, domain.Order{
			ID:     orderID,
			Status: domain.OrderDeclined,
		},
		domain.OrderAccepted,
	)
}

func (s *OrderService) GetOrderByID(ctx context.Context,
	orderID int) (*domain.Order, error) {
	return s.orderRepo.Get(ctx, orderID)
}
