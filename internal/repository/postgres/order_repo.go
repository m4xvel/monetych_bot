package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
}

func (r *OrderRepo) Create(ctx context.Context, order domain.Order) (int, error) {
	query := `
	INSERT INTO orders (user_id, status) 
	VALUES ($1, $2) RETURNING id` //RETURNING возвращает ID
	var id int
	if err := r.pool.QueryRow(ctx, query, order.UserID, order.Status).Scan(&id); err != nil {
		return 0, fmt.Errorf("insert order: %w", err)
	}
	return id, nil
}

func (r *OrderRepo) Accept(
	ctx context.Context,
	appraiserID int64,
	orderID int,
	status string,
	topicID, threadID int64,
) error {
	query := `
	UPDATE orders 
	SET appraiser_id = $1, status = $2, updated_at = now(), topic_id = $3, thread_id = $4
	WHERE id = $5`

	_, err := r.pool.Exec(ctx, query, appraiserID, status, topicID, threadID, orderID)
	if err != nil {
		return fmt.Errorf("Accept order: %w", err)
	}
	return nil
}

func (r *OrderRepo) GetByUser(ctx context.Context, userID int64) (*domain.Order, error) {
	query := `
	SELECT id, user_id, appraiser_id, status, topic_id, thread_id
	FROM orders 
	WHERE user_id = $1
	AND status = 'active'
	LIMIT 1`

	var o domain.Order
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&o.ID,
		&o.UserID,
		&o.AppraiserID,
		&o.Status,
		&o.TopicID,
		&o.ThreadID,
	)
	if err != nil {
		return nil, fmt.Errorf("get order status: %w", err)
	}
	return &o, nil
}

func (r *OrderRepo) GetByThread(ctx context.Context, topicID, threadID int64) (*domain.Order, error) {
	query := `
	SELECT id, user_id, appraiser_id, status, topic_id, thread_id
	FROM orders
	WHERE topic_id = $1 
	AND thread_id = $2
	AND status = 'active'
	LIMIT 1`

	var o domain.Order
	err := r.pool.QueryRow(ctx, query, topicID, threadID).Scan(
		&o.ID,
		&o.UserID,
		&o.AppraiserID,
		&o.Status,
		&o.TopicID,
		&o.ThreadID,
	)
	if err != nil {
		return nil, fmt.Errorf("get order status: %w", err)
	}
	return &o, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `
	UPDATE orders 
	SET status = $1, updated_at = now()
	WHERE id = $2`

	_, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}
