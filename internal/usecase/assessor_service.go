package usecase

import (
	"context"
	"fmt"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type AssessorService struct {
	repo domain.AssessorRepository
}

func NewAssessorService(repo domain.AssessorRepository) *AssessorService {
	return &AssessorService{repo: repo}
}

func (s *AssessorService) GetAllAssessorTgIDs(ctx context.Context) ([]int64, error) {
	allAssessors, err := s.repo.GetAllAssessor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list assessors: %w", err)
	}

	var tgIDs []int64
	for _, a := range allAssessors {
		tgIDs = append(tgIDs, a.TgID)
	}

	return tgIDs, nil
}

func (s *AssessorService) GetTopicIDByTgID(
	ctx context.Context,
	assessorID int64,
) int64 {
	topicID, err := s.repo.GetTopicIDByTgID(ctx, assessorID)
	if err != nil {
		return 0
	}
	return topicID
}
