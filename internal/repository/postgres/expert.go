package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

type ExpertRepo struct {
	pool *pgxpool.Pool
}

func NewExpertRepo(pool *pgxpool.Pool) *ExpertRepo {
	return &ExpertRepo{pool: pool}
}

func (r *ExpertRepo) Get(ctx context.Context) ([]domain.Expert, error) {
	const q = `
		SELECT id, topic_id, is_active
		FROM experts 
		WHERE is_active = true
	`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		logger.Log.Errorw("failed to query experts",
			"err", err,
		)
		return nil, err
	}
	defer rows.Close()

	var out []domain.Expert

	for rows.Next() {
		var e domain.Expert
		if err := rows.Scan(
			&e.ID,
			&e.TopicID,
			&e.IsActive,
		); err != nil {
			logger.Log.Errorw("failed to scan expert row",
				"err", err,
			)
			return nil, err
		}
		out = append(out, e)
	}

	if err := rows.Err(); err != nil {
		logger.Log.Errorw("rows error while iterating experts",
			"err", err,
		)
		return nil, err
	}

	return out, nil
}
