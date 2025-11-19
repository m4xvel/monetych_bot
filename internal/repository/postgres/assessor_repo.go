package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

type AssessorRepo struct {
	pool *pgxpool.Pool
}

func NewAssessorRepo(pool *pgxpool.Pool) *AssessorRepo {
	return &AssessorRepo{pool: pool}
}

func (r *AssessorRepo) GetByID(ctx context.Context, id int) (*domain.Assessor, error) {
	const q = `
	SELECT id, tg_id, topic_id
	FROM assessors
	WHERE id = $1
	LIMIT 1`

	var a domain.Assessor
	err := r.pool.QueryRow(ctx, q, id).Scan(&a.ID, &a.UserID, &a.TopicID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("assessor repo - GetByTgID: %w", err)
	}
	return &a, nil
}

func (r *AssessorRepo) GetByTgID(ctx context.Context, tgID int64) (*domain.Assessor, error) {
	const q = `
	SELECT id, tg_id, topic_id
	FROM assessors
	WHERE tg_id = $1
	LIMIT 1`

	var a domain.Assessor
	err := r.pool.QueryRow(ctx, q, tgID).Scan(&a.ID, &a.UserID, &a.TopicID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("assessor repo - GetByTgID: %w", err)
	}
	return &a, nil
}

func (r *AssessorRepo) GetAll(ctx context.Context) ([]domain.Assessor, error) {
	const q = `SELECT id, tg_id, topic_id FROM assessors`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("assessor repo - query GetAll: %w", err)
	}
	defer rows.Close()

	var out []domain.Assessor
	for rows.Next() {
		var a domain.Assessor
		if err := rows.Scan(&a.ID, &a.UserID, &a.TopicID); err != nil {
			return nil, fmt.Errorf("assessor repo - scan GetAll: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("assessor repo - rows err GetAll: %w", err)
	}
	return out, nil
}

func (r *AssessorRepo) GetTopicID(ctx context.Context, tgID int64) (int64, error) {
	const q = `SELECT topic_id FROM assessors WHERE tg_id = $1 LIMIT 1`
	var topicID sql.NullInt64
	err := r.pool.QueryRow(ctx, q, tgID).Scan(&topicID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("assessor repo - GetTopicID: %w", err)
	}
	if !topicID.Valid {
		return 0, nil
	}
	return topicID.Int64, nil
}
