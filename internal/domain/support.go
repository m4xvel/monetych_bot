package domain

import "context"

type Support struct {
	ID       int
	ChatID   int64
	ChatLink string
}

type SupportRepository interface {
	Get(ctx context.Context) (*Support, error)
}
