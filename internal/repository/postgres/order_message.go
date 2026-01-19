package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
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
	return err
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
		return nil, err
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
			return nil, err
		}

		result = append(result, om)
	}

	return result, rows.Err()
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

	return err
}
