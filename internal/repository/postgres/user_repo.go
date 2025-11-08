package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) AddUserIfNotExists(ctx context.Context, userID int64) error {
	query :=
		`INSERT INTO users (tg_id) 
	VALUES ($1) ON CONFLICT (tg_id) DO NOTHING`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *UserRepo) VerifyUser(ctx context.Context, userID int64) error {
	query := `UPDATE users SET is_verified = TRUE WHERE tg_id =$1`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *UserRepo) Is_verified(ctx context.Context, userID int64) bool {
	var isVerified bool
	err := r.pool.QueryRow(ctx, "SELECT is_verified FROM users WHERE tg_id=$1", userID).Scan(&isVerified)
	if err != nil {
		return false
	}
	return isVerified
}

func (r *UserRepo) ChangeTotalOrders(ctx context.Context, userID int64) {

}
