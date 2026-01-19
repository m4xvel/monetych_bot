package domain

import (
	"context"
	"time"
)

type OrderStatus string

const (
	OrderNew             OrderStatus = "new"
	OrderAccepted        OrderStatus = "accepted"
	OrderExpertConfirmed OrderStatus = "expert_confirmed"
	OrderCompleted       OrderStatus = "completed"
	OrderCanceled        OrderStatus = "canceled"
	OrderDeclined        OrderStatus = "declined"
)

type Order struct {
	ID         int
	Token      string
	UserID     int
	ExpertID   *int
	ThreadID   *int64
	Status     OrderStatus
	GameID     int
	GameTypeID int

	UserNameAtPurchase     string
	GameNameAtPurchase     string
	GameTypeNameAtPurchase string

	createdAt time.Time
	updatedAt time.Time

	UserChatID int64
	TopicID    *int64
}

type OrderRepository interface {
	Create(ctx context.Context, order Order) (int, error)
	UpdateStatus(ctx context.Context, order Order, status OrderStatus) error
	SetActive(ctx context.Context, order Order, status OrderStatus) error
	Get(ctx context.Context, orderID int) (*Order, error)
}
