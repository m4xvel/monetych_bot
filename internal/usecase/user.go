package usecase

import (
	"context"

	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

type UserService struct {
	users domain.UserRepository
}

func NewUserService(
	ur domain.UserRepository,
) *UserService {
	return &UserService{
		users: ur,
	}
}

func (s *UserService) AddUser(
	ctx context.Context,
	chatID int64,
	name string,
	getPhoto func() string,
) error {

	err := s.users.Add(
		ctx, domain.User{
			ChatID: chatID,
			Name:   name,
		},
	)

	if err != nil {
		if err == ErrUserAlreadyExists {
			logger.Log.Debugw("user already exists",
				"chat_id", chatID,
			)
			return nil
		}

		logger.Log.Errorw("failed to add user",
			"chat_id", chatID,
			"err", err,
		)
		return err
	}

	photoURL := getPhoto()
	if photoURL == "" {
		return nil
	}

	if err := s.users.UpdatePhoto(
		ctx,
		domain.User{
			PhotoURL: photoURL,
			ChatID:   chatID,
		},
	); err != nil {
		logger.Log.Errorw("failed to update user photo",
			"chat_id", chatID,
			"err", err,
		)

		return err
	}

	logger.Log.Infow("user photo updated",
		"chat_id", chatID,
	)

	return nil
}

func (s *UserService) GetByChatID(
	ctx context.Context,
	chatID int64,
) (*domain.User, error) {

	user, err := s.users.Get(ctx, domain.User{ChatID: chatID})
	if err != nil {
		logger.Log.Errorw("failed to get user by chat id",
			"chat_id", chatID,
			"err", err,
		)
		return nil, err
	}

	return user, nil
}
