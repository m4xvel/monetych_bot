package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) GetByID(ctx context.Context, id int) (*domain.User, error) {
	const q = `SELECT id, tg_id, is_verified, total_orders FROM users WHERE id = $1 LIMIT 1`
	var u domain.User
	var isVerified sql.NullBool
	var totalOrders sql.NullInt32
	err := r.pool.QueryRow(ctx, q, id).Scan(&u.ID, &u.UserID, &isVerified, &totalOrders)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("user repo - GetByID: %w", err)
	}
	if isVerified.Valid {
		u.IsVerified = isVerified.Bool
	}
	if totalOrders.Valid {
		u.TotalOrders = int(totalOrders.Int32)
	}
	return &u, nil
}

func (r *UserRepo) GetByTgID(ctx context.Context, tgID int64) (*domain.User, error) {
	const q = `SELECT id, tg_id, is_verified, total_orders FROM users WHERE tg_id = $1 LIMIT 1`
	var u domain.User
	var isVerified sql.NullBool
	var totalOrders sql.NullInt32
	err := r.pool.QueryRow(ctx, q, tgID).Scan(&u.ID, &u.UserID, &isVerified, &totalOrders)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("user repo - GetByTgID: %w", err)
	}
	if isVerified.Valid {
		u.IsVerified = isVerified.Bool
	}
	if totalOrders.Valid {
		u.TotalOrders = int(totalOrders.Int32)
	}
	return &u, nil
}

func (r *UserRepo) AddIfNotExists(ctx context.Context, tgID int64) error {
	const q = `
	INSERT INTO users (tg_id, created_at)
	VALUES ($1, now())
	ON CONFLICT (tg_id) DO NOTHING`
	_, err := r.pool.Exec(ctx, q, tgID)
	if err != nil {
		return fmt.Errorf("user repo - AddIfNotExists: %w", err)
	}
	return nil
}

func (r *UserRepo) Verify(ctx context.Context, tgID int64) error {
	const q = `UPDATE users SET is_verified = TRUE WHERE tg_id = $1`
	ct, err := r.pool.Exec(ctx, q, tgID)
	if err != nil {
		return fmt.Errorf("user repo - Verify: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepo) IsVerified(ctx context.Context, tgID int64) (bool, error) {
	const q = `SELECT is_verified FROM users WHERE tg_id = $1 LIMIT 1`
	var isVerified sql.NullBool
	err := r.pool.QueryRow(ctx, q, tgID).Scan(&isVerified)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, ErrNotFound
		}
		return false, fmt.Errorf("user repo - IsVerified: %w", err)
	}
	if !isVerified.Valid {
		return false, nil
	}
	return isVerified.Bool, nil
}

func (r *UserRepo) IncrementOrders(ctx context.Context, userID int) error {
	const q = `UPDATE users SET total_orders = total_orders + 1 WHERE id = $1`
	ct, err := r.pool.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("user repo - IncrementOrders: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
