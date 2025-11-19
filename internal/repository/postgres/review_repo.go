package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReviewRepo struct {
	pool *pgxpool.Pool
}

func NewReviewRepo(pool *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{pool: pool}
}

func (r *ReviewRepo) Create(ctx context.Context, orderID, userID, rate int) (int, error) {
	const q = `
	INSERT INTO reviews (order_id, user_id, rating, created_at)
	VALUES ($1, $2, $3, now())
	RETURNING id`
	var id int
	if err := r.pool.QueryRow(ctx, q, orderID, userID, rate).Scan(&id); err != nil {
		return 0, fmt.Errorf("review repo - Create: %w", err)
	}
	return id, nil
}

func (r *ReviewRepo) UpdateText(ctx context.Context, id int, text string) error {
	const q = `
	UPDATE reviews
	SET text = $1
	WHERE id = $2`
	ct, err := r.pool.Exec(ctx, q, text, id)
	if err != nil {
		return fmt.Errorf("review repo - UpdateText: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
