package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

type ReviewRepo struct {
	pool *pgxpool.Pool
}

func NewReviewRepo(pool *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{pool: pool}
}

func (r *ReviewRepo) Create(ctx context.Context,
	review domain.Review) error {
	const q = `
		INSERT INTO reviews (
			order_id, 
			rating
		)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`

	cmd, err := r.pool.Exec(ctx, q, review.OrderID, review.Rating)
	if err != nil {
		logger.Log.Errorw("failed to create review",
			"err", err,
		)
		return err
	}

	if cmd.RowsAffected() == 0 {
		return err
	}

	return nil
}

func (r *ReviewRepo) Set(
	ctx context.Context,
	review domain.Review,
	status domain.ReviewStatus,
) error {
	const q = `
			UPDATE reviews
			SET
				text = $2,
				status = $3
			WHERE id = $1
				AND status = $4
		`

	cmd, err := r.pool.Exec(
		ctx,
		q,
		review.ID,
		review.Text,
		review.Status,
		status,
	)
	if err != nil {
		logger.Log.Errorw("failed to update review",
			"err", err,
		)
		return err
	}

	if cmd.RowsAffected() == 0 {
		return err
	}

	return nil
}

func (r *ReviewRepo) Publish(
	ctx context.Context,
	reviewID int,
) error {
	const q = `
		UPDATE reviews
		SET
			status = $2,
			published_at = now()
		WHERE id = $1
	`

	cmd, err := r.pool.Exec(
		ctx,
		q,
		reviewID,
		domain.ReviewPublished,
	)
	if err != nil {
		logger.Log.Errorw("failed to publish review",
			"err", err,
		)
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
