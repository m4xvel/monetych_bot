package domain

import "context"

type UserRepository interface {
	AddUserIfNotExists(ctx context.Context, tgID int64) error
	Is_verified(ctx context.Context, userID int) bool
	ChangeTotalOrders(ctx context.Context, userID int)
}
