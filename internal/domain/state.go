package domain

import "context"

type UserState string

const (
	StateIdle          UserState = "idle"
	StateActiveOrder   UserState = "active_order"
	StateWritingReview UserState = "writing_review"
)

type State struct {
	UserID   int64
	State    UserState
	ReviewID *int
}

type UserStateRepo interface {
	Get(ctx context.Context, user User) (*State, error)
	Set(ctx context.Context, user User, state UserState, review Review) error
}
