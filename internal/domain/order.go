package domain

import (
	"context"
)

type OrderStatus string

const (
	OrderNew       OrderStatus = "new"
	OrderActive    OrderStatus = "active"
	OrderCompleted OrderStatus = "completed"
	OrderClosed    OrderStatus = "closed"
)

type Order struct {
	ID          int
	UserID      int
	AppraiserID *int
	Status      OrderStatus
	TopicID     *int64
	ThreadID    *int64
}

type OrderRepository interface {
	Create(ctx context.Context, userID int) (int, error)
	Get(ctx context.Context, id int) (*Order, error)
	GetByUser(ctx context.Context, userID int, status OrderStatus) (*Order, error)
	GetByThread(ctx context.Context, topicID, threadID int64) (*Order, error)

	Accept(ctx context.Context, orderID int, assessorID int, topicID, threadID int64) (*Order, error)

	AssignAssessor(ctx context.Context, orderID, assessorID int) error
	SetThread(ctx context.Context, orderID int, topicID, threadID int64) error
	UpdateStatus(ctx context.Context, orderID int, status OrderStatus) error
}
