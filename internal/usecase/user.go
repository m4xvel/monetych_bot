package usecase

import (
	"context"

	"github.com/m4xvel/monetych_bot/internal/domain"
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
			return nil
		}
		return err
	}

	photoURL := getPhoto()
	return s.users.UpdatePhoto(
		ctx, domain.User{
			PhotoURL: photoURL,
			ChatID:   chatID,
		},
	)
}

func (s *UserService) GetByChatID(ctx context.Context, chatID int64) (*domain.User, error) {
	return s.users.Get(ctx, domain.User{ChatID: chatID})
}
