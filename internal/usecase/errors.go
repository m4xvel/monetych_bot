package usecase

import "errors"

var (
	ErrOrderNotFound          = errors.New("order not found")
	ErrOrderInvalidTransition = errors.New("invalid order status transition")
	ErrOrderAlreadyExists     = errors.New("user already has active/new order")
	ErrAssessorNotFound       = errors.New("assessor not found")
	ErrOrderNotFinished       = errors.New("order is not finished")
	ErrInvalidRating          = errors.New("invalid rating value")
)
