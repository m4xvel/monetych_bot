package usecase

import (
	"context"
	"fmt"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type OrderService struct {
	orders    domain.OrderRepository
	users     domain.UserRepository
	assessors domain.AssessorRepository
}

func NewOrderService(
	or domain.OrderRepository,
	ur domain.UserRepository,
	ar domain.AssessorRepository,
) *OrderService {
	return &OrderService{
		orders:    or,
		users:     ur,
		assessors: ar,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID int) (int, error) {

	existing, err := s.orders.GetByUser(ctx, userID, domain.OrderNew)
	if err != nil {
		return 0, fmt.Errorf("create order: check existing: %w", err)
	}
	if existing != nil {
		return 0, ErrOrderAlreadyExists
	}

	existing, err = s.orders.GetByUser(ctx, userID, domain.OrderActive)
	if err != nil {
		return 0, fmt.Errorf("create order: check existing: %w", err)
	}
	if existing != nil {
		return 0, ErrOrderAlreadyExists
	}

	id, err := s.orders.Create(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("create order: %w", err)
	}
	return id, nil
}

func (s *OrderService) AcceptOrder(ctx context.Context, assessorID int, orderID int, topicID, threadID int64) error {
	_, err := s.assessors.GetByID(ctx, assessorID)
	if err != nil {
		return ErrAssessorNotFound
	}

	o, err := s.orders.Accept(ctx, orderID, assessorID, topicID, threadID)
	if err != nil {
		return fmt.Errorf("accept order: %w", err)
	}
	if o == nil {
		return ErrOrderInvalidTransition
	}
	return nil
}

func (s *OrderService) GetActiveByClient(ctx context.Context, userID int) (*domain.Order, error) {
	o, err := s.orders.GetByUser(ctx, userID, domain.OrderActive)
	if err != nil {
		return nil, fmt.Errorf("get active order by client: %w", err)
	}
	return o, nil
}

func (s *OrderService) GetActiveByThread(ctx context.Context, topicID, threadID int64) (*domain.Order, error) {
	o, err := s.orders.GetByThread(ctx, topicID, threadID)
	if err != nil {
		return nil, fmt.Errorf("get active order by thread: %w", err)
	}
	return o, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, id int) (*domain.Order, error) {
	o, err := s.orders.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get order by id: %w", err)
	}
	if o == nil {
		return nil, ErrOrderNotFound
	}
	return o, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int, newStatus domain.OrderStatus) error {
	o, err := s.orders.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("update status: fetch order: %w", err)
	}

	if o == nil {
		return ErrOrderNotFound
	}

	if !canChangeStatus(o.Status, newStatus) {
		return ErrOrderInvalidTransition
	}

	if err := s.orders.UpdateStatus(ctx, id, newStatus); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	if newStatus == domain.OrderCompleted {
		s.users.IncrementOrders(ctx, o.UserID)
	}
	return nil
}

func canChangeStatus(from, to domain.OrderStatus) bool {
	switch from {
	case domain.OrderNew:
		return to == domain.OrderActive
	case domain.OrderActive:
		return to == domain.OrderCompleted || to == domain.OrderClosed
	default:
		return false
	}
}
