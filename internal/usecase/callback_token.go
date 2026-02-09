package usecase

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/m4xvel/monetych_bot/internal/apperr"
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
		if errors.Is(err, apperr.ErrNotFound) {
			return apperr.Wrap(apperr.KindInvalid, "callback_token.consume", err)
		}
		return err
	}

	if err := json.Unmarshal(cb.Payload, dest); err != nil {
		return apperr.Wrap(apperr.KindInvalid, "callback_token.payload", err)
	}
	return nil
}

func (u *CallbackTokenService) DeleteByActionAndOrderID(
	ctx context.Context,
	action string,
	orderID int,
) error {
	return u.repo.DeleteByActionAndOrderID(ctx, action, orderID)
}
