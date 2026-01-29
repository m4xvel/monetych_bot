package usecase

import (
	"context"

	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
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
	err := s.repo.Set(
		ctx,
		domain.UserState{
			State: domain.StateIdle,
		},
		chatID,
	)

	if err != nil {
		logger.Log.Errorw("failed to set state idle",
			"chat_id", chatID,
			"err", err,
		)
		return err
	}

	logger.Log.Infow("state changed",
		"chat_id", chatID,
		"state", domain.StateIdle,
	)

	return nil
}

func (s *StateService) SetStateStart(
	ctx context.Context,
	chatID int64,
) error {
	err := s.repo.Set(
		ctx, domain.UserState{
			State: domain.StateStart,
		},
		chatID,
	)

	if err != nil {
		logger.Log.Errorw("failed to set state start",
			"chat_id", chatID,
			"err", err,
		)
		return err
	}

	logger.Log.Infow("state changed",
		"chat_id", chatID,
		"state", domain.StateStart,
	)

	return nil
}

func (s *StateService) SetStateCommunication(
	ctx context.Context,
	chatID int64,
	orderID *int,
) error {
	err := s.repo.Set(
		ctx,
		domain.UserState{
			State:   domain.StateCommunication,
			OrderID: orderID,
		},
		chatID,
	)

	if err != nil {
		logger.Log.Errorw("failed to set communication state",
			"chat_id", chatID,
			"order_id", orderID,
			"err", err,
		)
		return err
	}

	logger.Log.Infow("state changed",
		"chat_id", chatID,
		"state", domain.StateCommunication,
		"order_id", orderID,
	)

	return nil
}

func (s *StateService) SetStateWritingReview(
	ctx context.Context,
	chatID int64,
) error {
	err := s.repo.Set(
		ctx, domain.UserState{
			State: domain.StateWritingReview,
		},
		chatID,
	)

	if err != nil {
		logger.Log.Errorw("failed to set state writing review",
			"chat_id", chatID,
			"err", err,
		)
		return err
	}

	logger.Log.Infow("state changed",
		"chat_id", chatID,
		"state", domain.StateWritingReview,
	)

	return nil
}

func (s *StateService) GetStateByChatID(
	ctx context.Context,
	chatID int64,
) (*domain.UserState, error) {
	state, err := s.repo.GetByChatID(ctx, chatID)
	if err != nil {
		logger.Log.Errorw("failed to get user state by chat id",
			"chat_id", chatID,
			"err", err,
		)
		return nil, err
	}

	if state == nil {
		return &domain.UserState{
			State:      domain.StateIdle,
			UserChatID: &chatID,
		}, nil
	}

	return state, nil
}

func (s *StateService) GetStateByThreadID(
	ctx context.Context,
	threadID int64,
) (*domain.UserState, error) {
	state, err := s.repo.GetByThreadID(ctx, threadID)
	if err != nil {
		logger.Log.Errorw("failed to get user state by thread id",
			"thread_id", threadID,
			"err", err,
		)
		return nil, err
	}
	return state, nil
}
