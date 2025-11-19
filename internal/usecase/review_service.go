package usecase

import (
	"context"
	"fmt"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type ReviewService struct {
	reviews domain.ReviewRepository
	orders  *OrderService
}

func NewReviewService(
	rr domain.ReviewRepository,
	os *OrderService,
) *ReviewService {
	return &ReviewService{
		reviews: rr,
		orders:  os,
	}
}

func (s *ReviewService) AddReview(ctx context.Context, orderID, userID, rate int, text string) (int, error) {
	if rate < 1 || rate > 5 {
		return 0, ErrInvalidRating
	}
	o, err := s.orders.GetOrderByID(ctx, orderID)
	if err != nil {
		return 0, fmt.Errorf("add review: fetch order: %w", err)
	}

	if o.Status != domain.OrderCompleted {
		return 0, ErrOrderNotFinished
	}
	id, err := s.reviews.Create(ctx, orderID, userID, rate)
	if err != nil {
		return 0, fmt.Errorf("add review: create rating: %w", err)
	}
	if text != "" {
		if err := s.reviews.UpdateText(ctx, id, text); err != nil {
			return 0, fmt.Errorf("add review: set text: %w", err)
		}
	}
	return id, nil
}

func (s *ReviewService) UpdateText(ctx context.Context, reviewID int, text string) error {
	if text == "" {
		return fmt.Errorf("update review text: text cannot be empty")
	}
	if err := s.reviews.UpdateText(ctx, reviewID, text); err != nil {
		return fmt.Errorf("update review text: %w", err)
	}
	return nil
}
