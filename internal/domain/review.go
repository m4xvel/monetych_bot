package domain

import (
	"context"
)

type Review struct {
	ID int
}

type ReviewRepository interface {
	SetRate(ctx context.Context, order Order, user User, rate int) (int, error)
	UpdateText(ctx context.Context, text string, id int) error
}
