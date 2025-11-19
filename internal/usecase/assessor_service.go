package usecase

import (
	"context"
	"fmt"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type AssessorService struct {
	repo domain.AssessorRepository
}

func NewAssessorService(r domain.AssessorRepository) *AssessorService {
	return &AssessorService{repo: r}
}

func (s *AssessorService) GetByTgID(ctx context.Context, tgID int64) (*domain.Assessor, error) {
	a, err := s.repo.GetByTgID(ctx, tgID)
	if err != nil {
		return nil, fmt.Errorf("get assessor by tg id: %w", err)
	}
	return a, nil
}

func (s *AssessorService) GetAllTgIDs(ctx context.Context) ([]int64, error) {
	all, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all assessors: %w", err)
	}
	out := make([]int64, 0, len(all))
	for _, a := range all {
		out = append(out, a.UserID)
	}
	return out, nil
}
