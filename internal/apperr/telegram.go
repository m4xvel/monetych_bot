package apperr

import (
	"errors"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramError struct {
	Op         string
	Code       int
	Message    string
	Parameters *tgbotapi.ResponseParameters
	Err        error
}

func (e *TelegramError) Error() string {
	if e == nil {
		return ""
	}

	base := "telegram api error"
	if e.Code != 0 {
		base = fmt.Sprintf("telegram api error %d", e.Code)
	}
	if e.Message != "" {
		base = fmt.Sprintf("%s: %s", base, e.Message)
	}

	if e.Parameters != nil {
		parts := make([]string, 0, 2)
		if e.Parameters.RetryAfter != 0 {
			parts = append(parts, fmt.Sprintf("retry_after=%d", e.Parameters.RetryAfter))
		}
		if e.Parameters.MigrateToChatID != 0 {
			parts = append(parts, fmt.Sprintf("migrate_to_chat_id=%d", e.Parameters.MigrateToChatID))
		}
		if len(parts) > 0 {
			base = fmt.Sprintf("%s (%s)", base, strings.Join(parts, ", "))
		}
	}

	if e.Op != "" {
		base = fmt.Sprintf("%s: %s", e.Op, base)
	}

	if e.Err != nil && e.Message == "" {
		return fmt.Sprintf("%s: %v", base, e.Err)
	}

	return base
}

func (e *TelegramError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *TelegramError) Kind() Kind {
	switch e.Code {
	case 400:
		return KindInvalid
	case 401:
		return KindUnauthorized
	case 403:
		return KindForbidden
	case 404:
		return KindNotFound
	case 409:
		return KindConflict
	case 429:
		return KindRateLimited
	case 500, 502, 503, 504:
		return KindUnavailable
	default:
		return KindExternal
	}
}

func (e *TelegramError) Is(target error) bool {
	switch t := target.(type) {
	case *Error:
		if t.Kind == "" {
			return false
		}
		return e.Kind() == t.Kind
	case *TelegramError:
		if t.Code != 0 && e.Code != t.Code {
			return false
		}
		return true
	default:
		return false
	}
}

func WrapTelegram(op string, err error) error {
	if err == nil {
		return nil
	}

	var tgErr *tgbotapi.Error
	if errors.As(err, &tgErr) {
		params := tgErr.ResponseParameters
		return &TelegramError{
			Op:         op,
			Code:       tgErr.Code,
			Message:    tgErr.Message,
			Parameters: &params,
			Err:        err,
		}
	}

	return &TelegramError{
		Op:  op,
		Err: err,
	}
}
