package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

type OrderMessageRepo struct {
	pool *pgxpool.Pool
}

func NewOrderMessageRepo(pool *pgxpool.Pool) *OrderMessageRepo {
	return &OrderMessageRepo{pool: pool}
}

func (r *OrderMessageRepo) Save(ctx context.Context, orderMessage domain.OrderMessage) error {
	const q = `
		INSERT INTO order_messages (
			order_id,
			chat_id,
			message_id
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (order_id, chat_id, message_id)
		DO NOTHING;
	`

	_, err := r.pool.Exec(
		ctx,
		q,
		orderMessage.OrderID,
		orderMessage.ChatID,
		orderMessage.MessageID,
	)
	if err != nil {
		wrapped := dbErr("order_message.save", err)
		logger.Log.Errorw("failed to save order message",
			"err", wrapped,
		)
		return wrapped
	}

	return nil
}

func (r *OrderMessageRepo) Get(ctx context.Context, orderID int) ([]domain.OrderMessage, error) {
	const q = `
		SELECT
			id,
			order_id,
			chat_id,
			message_id,
			created_at,
			deleted_at
		FROM order_messages
		WHERE order_id = $1
			AND deleted_at IS NULL;
	`

	rows, err := r.pool.Query(ctx, q, orderID)
	if err != nil {
		wrapped := dbErr("order_message.get", err)
		logger.Log.Errorw("failed to query order messages",
			"err", wrapped,
		)
		return nil, wrapped
	}
	defer rows.Close()

	var result []domain.OrderMessage

	for rows.Next() {
		var om domain.OrderMessage

		if err := rows.Scan(
			&om.ID,
			&om.OrderID,
			&om.ChatID,
			&om.MessageID,
			&om.CreatedAt,
			&om.DeletedAt,
		); err != nil {
			wrapped := dbErr("order_message.scan", err)
			logger.Log.Errorw("failed to scan order message row",
				"err", wrapped,
			)
			return nil, wrapped
		}

		result = append(result, om)
	}

	if err := rows.Err(); err != nil {
		wrapped := dbErr("order_message.rows", err)
		logger.Log.Errorw("rows error while iterating order messages",
			"err", wrapped,
		)
		return nil, wrapped
	}

	return result, nil
}

func (r *OrderMessageRepo) Delete(ctx context.Context, orderID int) error {
	const q = `
		UPDATE order_messages
		SET deleted_at = $1
		WHERE order_id = $2
		  AND deleted_at IS NULL;
	`

	_, err := r.pool.Exec(
		ctx,
		q,
		time.Now(),
		orderID,
	)
	if err != nil {
		wrapped := dbErr("order_message.delete", err)
		logger.Log.Errorw("failed to mark order messages as deleted",
			"err", wrapped,
		)
		return wrapped
	}

	return nil
}

func (r *OrderMessageRepo) PurgeDeletedBefore(
	ctx context.Context,
	before time.Time,
) (int64, error) {
	const q = `
		DELETE FROM order_messages
		WHERE deleted_at IS NOT NULL
		  AND deleted_at < $1;
	`

	tag, err := r.pool.Exec(ctx, q, before)
	if err != nil {
		wrapped := dbErr("order_message.purge_deleted_before", err)
		logger.Log.Errorw("failed to purge deleted order messages",
			"before", before,
			"err", wrapped,
		)
		return 0, wrapped
	}

	return tag.RowsAffected(), nil
}
