package usecase

import (
	"context"
	"regexp"
	"strings"

	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
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
	orderID, err := s.orderRepo.Create(ctx, domain.Order{
		UserID:                 userID,
		GameID:                 gameID,
		GameTypeID:             gameTypeID,
		UserNameAtPurchase:     userNameAtPurchase,
		GameNameAtPurchase:     gameNameAtPurchase,
		GameTypeNameAtPurchase: gameTypeNameAtPurchase,
	})

	if err != nil {
		logger.Log.Errorw("failed to create order",
			"user_id", userID,
			"game_id", gameID,
			"game_type_id", gameTypeID,
			"err", err,
		)
		return 0, err
	}

	logger.Log.Infow("order created",
		"order_id", orderID,
		"user_id", userID,
		"game_id", gameID,
		"game_type_id", gameTypeID,
	)

	return orderID, nil
}

func (s *OrderService) SetExpertData(
	ctx context.Context,
	orderID int,
	expertID int,
	threadID int64,
) error {
	err := s.orderRepo.SetActive(
		ctx, domain.Order{
			ID:       orderID,
			ExpertID: &expertID,
			ThreadID: &threadID,
		},
		domain.OrderAccepted,
	)

	if err != nil {
		logger.Log.Errorw("failed to set expert data",
			"order_id", orderID,
			"expert_id", expertID,
			"err", err,
		)
		return err
	}

	logger.Log.Infow("expert assigned to order",
		"order_id", orderID,
		"expert_id", expertID,
	)

	return nil
}

func (s *OrderService) SetAcceptedStatus(ctx context.Context, orderID int) error {
	err := s.orderRepo.UpdateStatus(
		ctx,
		domain.Order{
			ID:     orderID,
			Status: domain.OrderAccepted,
		},
		domain.OrderNew,
	)

	if err != nil {
		logger.Log.Errorw("failed to accept order",
			"order_id", orderID,
			"from", domain.OrderNew,
			"to", domain.OrderAccepted,
			"err", err,
		)
		return err
	}

	logger.Log.Infow("order accepted",
		"order_id", orderID,
	)

	return nil
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
	if err := s.userRepo.IncrementOrders(ctx, chatID); err != nil {
		logger.Log.Errorw("failed to increment user orders",
			"user_chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
		return err
	}

	err := s.orderRepo.UpdateStatus(
		ctx,
		domain.Order{
			ID:     orderID,
			Status: domain.OrderCompleted,
		},
		domain.OrderExpertConfirmed,
	)

	if err != nil {
		logger.Log.Errorw("failed to complete order",
			"order_id", orderID,
			"err", err,
		)
		return err
	}

	logger.Log.Infow("order completed",
		"order_id", orderID,
		"user_chat_id", chatID,
	)

	return nil
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

func (s *OrderService) FindByToken(
	ctx context.Context,
	token string,
) (*domain.OrderFull, error) {

	t := regexp.MustCompile(`[^A-Z0-9]`).
		ReplaceAllString(strings.ToUpper(token), "")

	if len(t) != 12 {
		logger.Log.Warnw("invalid order token",
			"token_length", len(t),
		)
		return nil, ErrInvalidToken
	}

	return s.orderRepo.FindByToken(ctx, t)
}

func (s *OrderService) FindByID(
	ctx context.Context,
	orderID int,
) (*domain.OrderFull, error) {
	return s.orderRepo.FindByID(ctx, orderID)
}
