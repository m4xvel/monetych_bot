package domain

import "context"

type Assessor struct {
	ID         int
	TgID       int64
	OrdersDone int
}

type AssessorRepository interface {
	GetAllAssessor(ctx context.Context) ([]Assessor, error)
}
