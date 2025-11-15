package domain

import "context"

type Order struct {
	ID          int
	UserID      int64
	AppraiserID *int64
	Status      string
	TopicID     *int64
	ThreadID    *int64
}

type OrderRepository interface {
	Create(ctx context.Context, order Order) (int, error)
	Accept(ctx context.Context, appraiserID int64, orderID int, status string, topicID, threadID int64) error
	GetByID(ctx context.Context, id int) (*Order, error)
	GetByUser(ctx context.Context, userID int64, status string) (*Order, error)
	GetByThread(ctx context.Context, topicID, threadID int64) (*Order, error)
	UpdateStatus(ctx context.Context, id int, status string) error
}
