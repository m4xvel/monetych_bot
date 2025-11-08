package usecase

import (
	"context"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) VerifyUser(
	ctx context.Context, userID int64) {
	s.repo.VerifyUser(ctx, userID)
}

func (s *UserService) CheckStatusVerified(
	ctx context.Context, userID int64) bool {
	return s.repo.Is_verified(ctx, userID)
}

func (s *UserService) AddUserIfNotExists(
	ctx context.Context, userID int64) error {
	return s.repo.AddUserIfNotExists(ctx, userID)
}
