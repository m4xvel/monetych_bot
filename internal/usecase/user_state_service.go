package usecase

import (
	"context"
	"fmt"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type StateService struct {
	repo domain.UserStateRepo
}

func NewStateService(r domain.UserStateRepo) *StateService {
	return &StateService{repo: r}
}

func (s *StateService) GetState(ctx context.Context, userID int) (*domain.UserState, error) {
	st, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get state: %w", err)
	}
	return st, nil
}

func (s *StateService) SetState(ctx context.Context, state domain.UserState) error {
	return s.repo.Set(ctx, state)
}
