package domain

import "context"

type Order struct {
	ID          int
	UserID      int64
	AppraiserID int64
	Status      string
}

type OrderRepository interface {
	Create(ctx context.Context, order Order) (int, error)
	Accept(ctx context.Context, appraiserID int64, orderID int, status string) error
	GetByUser(ctx context.Context, userID int64) (*Order, error)
	UpdateStatus(ctx context.Context, id int, status string) error
}
