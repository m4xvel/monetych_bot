package usecase

import (
	"context"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type StateService struct {
	repo domain.UserStateRepository
}

func NewStateService(r domain.UserStateRepository) *StateService {
	return &StateService{repo: r}
}

func (s *StateService) SetStateIdle(
	ctx context.Context,
	chatID int64,
) error {
	return s.repo.Set(
		ctx, domain.UserState{
			State: domain.StateIdle,
		},
		chatID,
	)
}

func (s *StateService) SetStateStart(
	ctx context.Context,
	chatID int64,
) error {
	return s.repo.Set(
		ctx, domain.UserState{
			State: domain.StateStart,
		},
		chatID,
	)
}

func (s *StateService) SetStateCommunication(
	ctx context.Context,
	chatID int64,
	orderID *int,
) error {
	return s.repo.Set(
		ctx, domain.UserState{
			State:   domain.StateCommunication,
			OrderID: orderID,
		},
		chatID,
	)
}

func (s *StateService) SetStateWritingReview(
	ctx context.Context,
	chatID int64,
) error {
	return s.repo.Set(
		ctx, domain.UserState{
			State: domain.StateWritingReview,
		},
		chatID,
	)
}

func (s *StateService) GetStateByChatID(
	ctx context.Context,
	chatID int64,
) (*domain.UserState, error) {
	return s.repo.GetByChatID(ctx, chatID)
}

func (s *StateService) GetStateByThreadID(
	ctx context.Context,
	threadID int64,
) (*domain.UserState, error) {
	return s.repo.GetByThreadID(ctx, threadID)
}
