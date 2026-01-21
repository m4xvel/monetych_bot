package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

type SupportRepo struct {
	pool *pgxpool.Pool
}

func NewSupportRepo(pool *pgxpool.Pool) *SupportRepo {
	return &SupportRepo{pool: pool}
}

func (r *SupportRepo) Get(ctx context.Context) (*domain.Support, error) {
	const q = `
		SELECT 
			id,
			chat_id,
			chat_link 
		FROM support
	`

	var s domain.Support
	err := r.pool.QueryRow(ctx, q).Scan(&s.ID, &s.ChatID, &s.ChatLink)

	if err != nil {
		return nil, err
	}

	if err == pgx.ErrNoRows {
		return nil, err
	}

	return &s, nil
}
