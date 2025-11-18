package usecase

import (
	"context"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type StateService struct {
	repo domain.UserStateRepo
}

func NewStateService(repo domain.UserStateRepo) *StateService {
	return &StateService{repo: repo}
}

func (s *StateService) GetState(ctx context.Context, userID int) (*domain.State, error) {
	return s.repo.Get(ctx, domain.User{ID: userID})
}

func (s *StateService) SetState(ctx context.Context, userID int, state domain.UserState, reviewID int) error {
	return s.repo.Set(ctx, domain.User{ID: userID}, state, domain.Review{ID: reviewID})
}
