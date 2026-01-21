package domain

import (
	"context"
	"time"
)

type ReviewStatus string

const (
	ReviewRated     ReviewStatus = "rated"
	ReviewWithText  ReviewStatus = "with_text"
	ReviewPublished ReviewStatus = "published"
	ReviewRejected  ReviewStatus = "rejected"
)

type Review struct {
	ID          int
	OrderID     int
	Rating      int
	Text        *string
	Status      ReviewStatus
	CreatedAt   *time.Time
	PublishedAt *time.Time
}

type ReviewRepository interface {
	Create(ctx context.Context, review Review) error
	Set(ctx context.Context, review Review, status ReviewStatus) error
	Publish(ctx context.Context, reviewID int) error
}
