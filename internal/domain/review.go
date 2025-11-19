package domain

import (
	"context"
)

type Review struct {
	ID int
}

type ReviewRepository interface {
	Create(ctx context.Context, orderID, userID, rate int) (int, error)
	UpdateText(ctx context.Context, id int, text string) error
}
