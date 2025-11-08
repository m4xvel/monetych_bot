package domain

import "context"

type UserRepository interface {
	AddUserIfNotExists(ctx context.Context, userID int64) error
	VerifyUser(ctx context.Context, userID int64) error
	Is_verified(ctx context.Context, userID int64) bool
	ChangeTotalOrders(ctx context.Context, userID int64)
}
