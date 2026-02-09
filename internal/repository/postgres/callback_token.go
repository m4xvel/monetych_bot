package postgres

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

type CallbackTokenRepo struct {
	pool *pgxpool.Pool
}

func NewCallbackTokenRepo(pool *pgxpool.Pool) *CallbackTokenRepo {
	return &CallbackTokenRepo{pool: pool}
}

func (r *CallbackTokenRepo) Create(
	ctx context.Context,
	callback *domain.CallbackToken,
) error {
	const q = `
		INSERT INTO callback_tokens (token, action, payload)
		VALUES ($1, $2, $3)
	`

	if _, err := r.pool.Exec(
		ctx, q,
		callback.Token,
		callback.Action,
		callback.Payload,
	); err != nil {
		wrapped := dbErr("callback_token.create", err)
		logger.Log.Errorw("failed to create callback token",
			"err", wrapped,
		)
		return wrapped
	}

	return nil
}

func (r *CallbackTokenRepo) Consume(
	ctx context.Context,
	callback *domain.CallbackToken,
) error {
	const q = `
		DELETE FROM callback_tokens
		WHERE token=$1 AND action=$2
		RETURNING payload
	`

	err := r.pool.QueryRow(
		ctx, q,
		callback.Token,
		callback.Action,
	).Scan(&callback.Payload)

	if errors.Is(err, pgx.ErrNoRows) {
		wrapped := dbErr("callback_token.consume", err)
		logger.Log.Errorw("error no rows callback token",
			"err", wrapped,
		)
		return wrapped
	}

	if err != nil {
		wrapped := dbErr("callback_token.consume", err)
		logger.Log.Errorw("failed to consume callback token",
			"err", wrapped,
		)
		return wrapped
	}

	return nil
}

func (r *CallbackTokenRepo) DeleteByActionAndOrderID(
	ctx context.Context,
	action string,
	orderID int,
) error {

	const q = `
		DELETE FROM callback_tokens
		WHERE action = $1
		  AND payload->>'order_id' = $2
	`

	_, err := r.pool.Exec(
		ctx,
		q,
		action,
		strconv.Itoa(orderID),
	)

	if err != nil {
		return dbErr("callback_token.delete_by_action_and_order_id", err)
	}
	return nil
}
