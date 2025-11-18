package domain

import "context"

type User struct {
	ID     int
	UserID int64
}

type UserRepository interface {
	GetUserByUserID(ctx context.Context, id int) (*User, error)
	GetUserByUserTgID(ctx context.Context, userID int64) (*User, error)
	AddUserIfNotExists(ctx context.Context, userID int64) error
	VerifyUser(ctx context.Context, userID int64) error
	Is_verified(ctx context.Context, userID int64) bool
	ChangeTotalOrders(ctx context.Context, userID int64)
}
