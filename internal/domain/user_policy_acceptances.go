package domain

import (
	"context"
	"time"
)

type UserPolicyAcceptances struct {
	ID         int
	UserID     int
	PolicyID   int
	AcceptedAt time.Time
}

type UserPolicyAcceptancesRepository interface {
	Set(ctx context.Context, chatID int64, titles []string) error
	IsUserAccepted(ctx context.Context, chatID int64, titles []string) (bool, error)
}
