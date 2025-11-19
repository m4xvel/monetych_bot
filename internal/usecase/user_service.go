package usecase

import (
	"context"
	"fmt"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type UserService struct {
	users domain.UserRepository
}

func NewUserService(ur domain.UserRepository) *UserService {
	return &UserService{users: ur}
}

func (s *UserService) GetByID(ctx context.Context, id int) (*domain.User, error) {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (s *UserService) GetByTgID(ctx context.Context, tgID int64) (*domain.User, error) {
	u, err := s.users.GetByTgID(ctx, tgID)
	if err != nil {
		return nil, fmt.Errorf("get user by tg id: %w", err)
	}
	return u, nil
}

func (s *UserService) AddIfNotExists(ctx context.Context, tgID int64) error {
	return s.users.AddIfNotExists(ctx, tgID)
}

func (s *UserService) Verify(ctx context.Context, tgID int64) error {
	return s.users.Verify(ctx, tgID)
}

func (s *UserService) IsVerified(ctx context.Context, tgID int64) (bool, error) {
	return s.users.IsVerified(ctx, tgID)
}
