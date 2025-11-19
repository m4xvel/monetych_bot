package domain

import "context"

type User struct {
	ID          int
	UserID      int64
	IsVerified  bool
	TotalOrders int
}

type UserRepository interface {
	GetByID(ctx context.Context, id int) (*User, error)
	GetByTgID(ctx context.Context, tgID int64) (*User, error)

	AddIfNotExists(ctx context.Context, tgID int64) error
	Verify(ctx context.Context, tgID int64) error
	IsVerified(ctx context.Context, tgID int64) (bool, error)

	IncrementOrders(ctx context.Context, userID int) error
}
