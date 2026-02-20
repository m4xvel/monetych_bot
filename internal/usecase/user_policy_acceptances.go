package usecase

import (
	"context"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type UserPolicyAcceptancesService struct {
	repo           domain.UserPolicyAcceptancesRepository
	requiredTitles []string
}

func NewUserPolicyAcceptancesService(
	r domain.UserPolicyAcceptancesRepository,
	privacyPolicyTitle string,
	publicOfferTitle string,
) *UserPolicyAcceptancesService {
	return &UserPolicyAcceptancesService{
		repo: r,
		requiredTitles: []string{
			privacyPolicyTitle,
			publicOfferTitle,
		},
	}
}

func (s *UserPolicyAcceptancesService) Accept(
	ctx context.Context,
	chatID int64,
) error {
	return s.repo.Set(ctx, chatID, s.requiredTitles)
}

func (uc *UserPolicyAcceptancesService) IsAccepted(
	ctx context.Context,
	chatID int64,
) (bool, error) {
	return uc.repo.IsUserAccepted(ctx, chatID, uc.requiredTitles)
}
