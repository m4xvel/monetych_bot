package domain

import (
	"context"
)

type Assessor struct {
	ID         int
	TgID       int64
	OrdersDone int
	TopicID    int64
}

type AssessorRepository interface {
	GetAllAssessor(ctx context.Context) ([]Assessor, error)
	GetTopicIDByTgID(ctx context.Context, tgID int64) (int64, error)
}
