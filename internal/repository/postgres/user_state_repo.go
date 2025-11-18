package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

type UserStateRepo struct {
	pool *pgxpool.Pool
}

func NewUserStateRepo(pool *pgxpool.Pool) *UserStateRepo {
	return &UserStateRepo{pool: pool}
}

func (r *UserStateRepo) Get(ctx context.Context, user domain.User) (*domain.State, error) {
	query := `
	SELECT state, review_id
	FROM user_state
	WHERE user_id = $1`

	var s domain.State
	err := r.pool.QueryRow(ctx, query, user.ID).Scan(&s.State, &s.ReviewID)
	if err != nil {
		return nil, fmt.Errorf("get user state: %s", err)
	}
	return &s, nil
}

func (r *UserStateRepo) Set(ctx context.Context, user domain.User, state domain.UserState, review domain.Review) error {
	query := `
	INSERT INTO user_state (user_id, state, review_id)
	VALUES ($1, $2, $3)
	ON CONFLICT (user_id) DO UPDATE SET state = EXCLUDED.state, updated_at = NOW()`

	_, err := r.pool.Exec(ctx, query, user.ID, state, review.ID)
	if err != nil {
		return fmt.Errorf("set user state: %s", err)
	}
	return nil
}
