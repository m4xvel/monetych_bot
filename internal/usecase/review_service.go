package usecase

import (
	"context"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type ReviewService struct {
	repo domain.ReviewRepository
}

func NewReviewService(repo domain.ReviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

func (s *ReviewService) SetRate(ctx context.Context, orderID, userID, rate int) (int, error) {
	return s.repo.SetRate(ctx, domain.Order{ID: orderID}, domain.User{ID: userID}, rate)
}

func (s *ReviewService) UpdateText(ctx context.Context, text string, id int) error {
	return s.repo.UpdateText(ctx, text, id)
}
