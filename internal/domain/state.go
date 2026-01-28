package domain

import (
	"context"
	"time"
)

type StateName string

const (
	StateIdle          StateName = "idle"
	StateStart         StateName = "start"
	StateCommunication StateName = "communication"
	StateWritingReview StateName = "writing_review"
)

type UserState struct {
	State     StateName
	OrderID   *int
	UpdatedAt time.Time

	ExpertID      *int
	ExpertTopicID *int64

	OrderThreadID *int64
	OrderStatus   *OrderStatus

	UserID     *int
	UserChatID *int64

	ReviewID *int
}

type UserStateRepository interface {
	Set(ctx context.Context, state UserState, chatID int64) error
	GetByChatID(ctx context.Context, chatID int64) (*UserState, error)
	GetByThreadID(ctx context.Context, threadID int64) (*UserState, error)
}
