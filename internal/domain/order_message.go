package domain

import (
	"context"
	"time"
)

type OrderMessage struct {
	ID        int
	OrderID   int
	ChatID    int64
	MessageID int
	CreatedAt time.Time
	DeletedAt *time.Time
}

type OrderMessageRepository interface {
	Save(ctx context.Context, orderMessage OrderMessage) error
	Get(ctx context.Context, orderID int) ([]OrderMessage, error)
	Delete(ctx context.Context, orderID int) error
	PurgeDeletedBefore(ctx context.Context, before time.Time) (int64, error)
}
