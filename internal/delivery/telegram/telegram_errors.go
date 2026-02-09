package telegram

import (
	"errors"

	"github.com/m4xvel/monetych_bot/internal/apperr"
)

func wrapTelegramErr(op string, err error) error {
	return apperr.WrapTelegram(op, err)
}

func isOrderAlreadyProcessed(err error) bool {
	var dbErr *apperr.DBError
	if errors.As(err, &dbErr) {
		return dbErr.Code == apperr.DBCodeOrderAlreadyProcessed
	}
	return false
}

func isInvalidToken(err error) bool {
	return errors.Is(err, apperr.ErrInvalid)
}
