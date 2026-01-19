package usecase

import (
	"context"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type ReviewService struct {
	repo domain.ReviewRepository
}

func NewReviewService(r domain.ReviewRepository) *ReviewService {
	return &ReviewService{repo: r}
}

func (r *ReviewService) Rate(
	ctx context.Context,
	orderID int,
	rating int,
) error {
	return r.repo.Create(
		ctx, domain.Review{
			OrderID: orderID,
			Rating:  rating,
		},
	)
}

func (r *ReviewService) AddText(
	ctx context.Context,
	reviewID int,
	text string,
) error {
	return r.repo.Set(
		ctx, domain.Review{
			ID:     reviewID,
			Text:   &text,
			Status: domain.ReviewWithText,
		},
		domain.ReviewRated,
	)
}
