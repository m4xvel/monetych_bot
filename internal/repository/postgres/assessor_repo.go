package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

type AssessorRepo struct {
	pool *pgxpool.Pool
}

func NewAssessorRepo(pool *pgxpool.Pool) *AssessorRepo {
	return &AssessorRepo{pool: pool}
}

func (r *AssessorRepo) GetAllAssessor(ctx context.Context) ([]domain.Assessor, error) {
	rows, err := r.pool.Query(ctx, `
	SELECT id, tg_id, orders_done, topic_id 
	FROM assessors`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assessors []domain.Assessor
	for rows.Next() {
		var a domain.Assessor
		if err := rows.Scan(&a.ID, &a.TgID, &a.OrdersDone, &a.TopicID); err != nil {
			return nil, err
		}
		assessors = append(assessors, a)
	}

	return assessors, nil
}

func (r *AssessorRepo) GetTopicIDByTgID(ctx context.Context, tgID int64) (int64, error) {
	query := `
	SELECT topic_id
	FROM assessors
	WHERE tg_id = $1`

	var topicID int64
	if err := r.pool.QueryRow(ctx, query, tgID).Scan(&topicID); err != nil {
		return 0, fmt.Errorf("select topic: %w", err)
	}
	return topicID, nil
}
