package usecase

import (
	"context"

	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
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

	err := r.repo.Create(
		ctx,
		domain.Review{
			OrderID: orderID,
			Rating:  rating,
		},
	)

	if err != nil {
		logger.Log.Errorw("failed to create review",
			"order_id", orderID,
			"rating", rating,
			"err", err,
		)
		return err
	}

	logger.Log.Infow("review rated",
		"order_id", orderID,
		"rating", rating,
	)

	return nil
}

func (r *ReviewService) AddText(
	ctx context.Context,
	reviewID int,
	text string,
) error {

	err := r.repo.Set(
		ctx,
		domain.Review{
			ID:     reviewID,
			Text:   &text,
			Status: domain.ReviewWithText,
		},
		domain.ReviewRated,
	)

	if err != nil {
		logger.Log.Errorw("failed to add review text",
			"review_id", reviewID,
			"err", err,
		)
		return err
	}

	logger.Log.Infow("review text added",
		"review_id", reviewID,
	)

	return nil
}

func (r *ReviewService) Publish(
	ctx context.Context,
	reviewID int,
) error {

	err := r.repo.Publish(ctx, reviewID)
	if err != nil {
		logger.Log.Errorw("failed to publish review",
			"review_id", reviewID,
			"err", err,
		)
		return err
	}

	logger.Log.Infow("review published",
		"review_id", reviewID,
	)

	return nil
}
