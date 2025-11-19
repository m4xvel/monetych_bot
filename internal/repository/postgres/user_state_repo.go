package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

type UserStateRepo struct {
	pool *pgxpool.Pool
}

func NewUserStateRepo(pool *pgxpool.Pool) *UserStateRepo {
	return &UserStateRepo{pool: pool}
}

func (r *UserStateRepo) Get(ctx context.Context, userID int) (*domain.UserState, error) {
	const q = `SELECT user_id, state, review_id, updated_at FROM user_state WHERE user_id = $1 LIMIT 1`
	var s domain.UserState
	var review sql.NullInt32
	var updatedAt sql.NullTime

	err := r.pool.QueryRow(ctx, q, userID).Scan(&s.UserID, &s.State, &review, &updatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("user state repo - Get: %w", err)
	}
	if review.Valid {
		v := int(review.Int32)
		s.ReviewID = &v
	} else {
		s.ReviewID = nil
	}
	if updatedAt.Valid {
		if t, ok := any(&s).(interface {
			SetUpdatedAt(time.Time)
		}); ok {
			t.SetUpdatedAt(updatedAt.Time)
		}
	}
	return &s, nil
}

func (r *UserStateRepo) Set(ctx context.Context, s domain.UserState) error {
	const q = `
	INSERT INTO user_state (user_id, state, review_id, updated_at)
	VALUES ($1, $2, $3, now())
	ON CONFLICT (user_id)
	DO UPDATE SET state = EXCLUDED.state, review_id = EXCLUDED.review_id, updated_at = now()`
	var review interface{}
	if s.ReviewID == nil {
		review = nil
	} else {
		review = *s.ReviewID
	}
	_, err := r.pool.Exec(ctx, q, s.UserID, s.State, review)
	if err != nil {
		return fmt.Errorf("user state repo - Set: %w", err)
	}
	return nil
}
