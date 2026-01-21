package domain

import "context"

type Expert struct {
	ID       int
	ChatID   int64
	TopicID  int64
	IsActive bool
}

type ExpertRepository interface {
	Get(ctx context.Context) ([]Expert, error)
}
