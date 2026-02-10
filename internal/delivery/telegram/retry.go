package telegram

import (
	"errors"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

const maxTelegramRetryAttempts = 3

func retryOnRateLimit(op string, fn func() error, fields ...any) error {
	return retryOnRateLimitWithAttempts(op, maxTelegramRetryAttempts, fn, fields...)
}

func retryOnRateLimitForever(op string, fn func() error, fields ...any) error {
	return retryOnRateLimitWithAttempts(op, 0, fn, fields...)
}

func retryOnRateLimitWithAttempts(
	op string,
	maxAttempts int,
	fn func() error,
	fields ...any,
) error {
	var lastErr error

	for attempt := 1; maxAttempts <= 0 || attempt <= maxAttempts; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			retryAfter, ok := retryAfterSeconds(err)
			if !ok || (maxAttempts > 0 && attempt == maxAttempts) {
				return lastErr
			}

			keyvals := []any{
				"op", op,
				"retry_after", retryAfter,
				"attempt", attempt,
			}
			if len(fields) > 0 {
				keyvals = append(keyvals, fields...)
			}

			logger.Log.Warnw("rate limited, retrying request", keyvals...)
			time.Sleep(time.Duration(retryAfter) * time.Second)
		}
	}

	return lastErr
}

func retryAfterSeconds(err error) (int, bool) {
	var tgErr *tgbotapi.Error
	if errors.As(err, &tgErr) && tgErr != nil && tgErr.Code == 429 && tgErr.RetryAfter > 0 {
		return tgErr.RetryAfter, true
	}
	return 0, false
}
