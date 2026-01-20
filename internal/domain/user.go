package domain

import (
	"context"
	"time"
)

type User struct {
	ID          int
	ChatID      int64
	Name        string
	PhotoURL    string
	IsVerified  bool
	CreatedAt   time.Time
	TotalOrders int
}

type UserRepository interface {
	Add(ctx context.Context, user User) error
	UpdatePhoto(ctx context.Context, user User) error
	Get(ctx context.Context, user User) (*User, error)
	IncrementOrders(ctx context.Context, chatID int64) error
}
