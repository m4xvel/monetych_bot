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
	INSERT INTO orders (user_id, appraiser_id, status) 
	VALUES ($1, $2, $3) RETURNING id` //RETURNING возвращает ID
	var id int
	if err := r.pool.QueryRow(ctx, query, order.ID, order.AppraiserID, order.Status).Scan(&id); err != nil {
		return 0, fmt.Errorf("insert order: %w", err)
	}
	return id, nil
}

func (r *OrderRepo) GetByUser(ctx context.Context, userID int64) (*domain.Order, error) {
	query := `
	SELECT (id, user_id, appraiser_id, status) 
	FROM orders 
	WHERE (user_id = $1 OR appraiser_id= $1)
	AND status = 'active'
	LIMIT 1`

	var d domain.Order
	if err := r.pool.QueryRow(ctx, query, userID).Scan(
		&d.ID, &d.UserID, &d.AppraiserID, &d.Status,
	); err != nil {
		return nil, fmt.Errorf("get by user: %w", err)
	}
	return &d, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `
	UPDATE orders 
	SET status = $1, updated_at = now()
	WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}
