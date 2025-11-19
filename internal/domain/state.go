package domain

import (
	"context"
	"time"
)

type StateName string

const (
	StateIdle          StateName = "idle"
	StateStart         StateName = "start"
	StateWritingReview StateName = "writing_review"
)

type UserState struct {
	UserID    int
	State     StateName
	ReviewID  *int
	UpdatedAt time.Time
}

type UserStateRepo interface {
	Get(ctx context.Context, userID int) (*UserState, error)
	Set(ctx context.Context, state UserState) error
}
