package domain

import "context"

type Order struct {
	ID          int
	UserID      int
	AppraiserID *int
	Status      string
	TopicID     *int64
	ThreadID    *int64
}

type OrderRepository interface {
	Create(ctx context.Context, user User, order Order) (int, error)
	Accept(ctx context.Context, assessor Assessor, orderID int, status string, topicID, threadID int64) error
	GetByID(ctx context.Context, id int) (*Order, error)
	GetByUser(ctx context.Context, user User, status string) (*Order, error)
	GetByThread(ctx context.Context, topicID, threadID int64) (*Order, error)
	UpdateStatus(ctx context.Context, id int, status string) error
}
