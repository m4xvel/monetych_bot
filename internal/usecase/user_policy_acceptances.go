package usecase

import (
	"context"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type UserPolicyAcceptancesService struct {
	repo domain.UserPolicyAcceptancesRepository
}

func NewUserPolicyAcceptancesService(
	r domain.UserPolicyAcceptancesRepository,
) *UserPolicyAcceptancesService {
	return &UserPolicyAcceptancesService{
		repo: r,
	}
}

const CurrentPolicyVersion = "1.0"

func (s *UserPolicyAcceptancesService) Accept(
	ctx context.Context,
	chatID int64,
) error {
	return s.repo.Set(ctx, chatID, CurrentPolicyVersion)
}

func (uc *UserPolicyAcceptancesService) IsAccepted(
	ctx context.Context,
	chatID int64,
) (bool, error) {
	return uc.repo.IsUserAccepted(ctx, chatID, CurrentPolicyVersion)
}
