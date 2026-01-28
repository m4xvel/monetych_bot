package usecase

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

type CallbackTokenService struct {
	repo domain.CallbackTokenRepository
}

func NewCallbackTokenService(
	repo domain.CallbackTokenRepository,
) *CallbackTokenService {
	return &CallbackTokenService{repo: repo}
}

func (s *CallbackTokenService) Create(
	ctx context.Context,
	action string,
	payload any,
) (string, error) {
	token := uuid.NewString()

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	err = s.repo.Create(ctx, &domain.CallbackToken{
		Token:   token,
		Action:  action,
		Payload: data,
	})
	if err != nil {
		return "", err
	}

	return token, nil
}

func (u *CallbackTokenService) Consume(
	ctx context.Context,
	token string,
	action string,
	dest any,
) error {
	cb := &domain.CallbackToken{
		Token:  token,
		Action: action,
	}

	if err := u.repo.Consume(ctx, cb); err != nil {
		return err
	}

	return json.Unmarshal(cb.Payload, dest)
}

func (u *CallbackTokenService) DeleteByActionAndOrderID(
	ctx context.Context,
	action string,
	orderID int,
) error {
	return u.repo.DeleteByActionAndOrderID(ctx, action, orderID)
}
