package domain

import (
	"context"
)

type Assessor struct {
	ID      int
	UserID  int64
	TopicID int64
}

type AssessorRepository interface {
	GetByID(ctx context.Context, id int) (*Assessor, error)
	GetByTgID(ctx context.Context, tgID int64) (*Assessor, error)
	GetAll(ctx context.Context) ([]Assessor, error)
	GetTopicID(ctx context.Context, tgID int64) (int64, error)
}
