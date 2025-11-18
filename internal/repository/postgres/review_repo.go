package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

type ReviewRepo struct {
	pool *pgxpool.Pool
}

func NewReviewRepo(pool *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{pool: pool}
}

func (r *ReviewRepo) SetRate(ctx context.Context, order domain.Order, user domain.User, rate int) (int, error) {
	query := `
	INSERT INTO reviews (order_id, user_id, rating)
	VALUES ($1, $2, $3) RETURNING id`
	var id int
	if err := r.pool.QueryRow(ctx, query, order.ID, user.ID, rate).Scan(&id); err != nil {
		return 0, fmt.Errorf("insert review: %w", err)
	}
	return id, nil
}

func (r *ReviewRepo) UpdateText(ctx context.Context, text string, id int) error {
	query := `
	UPDATE reviews
	SET text = $1
	WHERE id = $2`

	if _, err := r.pool.Exec(ctx, query, text, id); err != nil {
		return fmt.Errorf("insert review: %w", err)
	}
	return nil
}
