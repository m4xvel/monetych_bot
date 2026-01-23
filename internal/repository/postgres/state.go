package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

type UserStateRepo struct {
	pool *pgxpool.Pool
}

func NewUserStateRepo(pool *pgxpool.Pool) *UserStateRepo {
	return &UserStateRepo{pool: pool}
}

func (r *UserStateRepo) Set(
	ctx context.Context,
	state domain.UserState,
	chatID int64,
) error {
	const q = `
		INSERT INTO user_state (
			user_id, 
			state, 
			order_id, 
			updated_at
			)
		SELECT u.id, $2, $3, NOW()
		FROM users u
		WHERE u.chat_id = $1
		ON CONFLICT (user_id)
		DO UPDATE SET
			state = EXCLUDED.state,
			order_id = COALESCE(EXCLUDED.order_id, user_state.order_id),
			updated_at = NOW()
	`

	_, err := r.pool.Exec(ctx, q, chatID, state.State, state.OrderID)
	if err != nil {
		logger.Log.Errorw("failed to set user state",
			"err", err,
		)
		return err
	}

	return nil
}

func (r *UserStateRepo) GetByChatID(
	ctx context.Context,
	chatID int64,
) (*domain.UserState, error) {
	const q = `
			SELECT
				us.state,
				us.order_id,
				e.topic_id,
				o.thread_id,
				r.id,
				u.id,
				u.chat_id
			FROM users u
			JOIN user_state us
    		ON us.user_id = u.id
			LEFT JOIN orders o
				ON o.id = us.order_id
			LEFT JOIN experts e
				ON e.id = o.expert_id
			LEFT JOIN reviews r
				ON r.order_id = o.id
			WHERE u.chat_id = $1
		`

	var us domain.UserState

	err := r.pool.QueryRow(ctx, q, chatID).Scan(
		&us.State,
		&us.OrderID,
		&us.ExpertTopicID,
		&us.OrderThreadID,
		&us.ReviewID,
		&us.UserID,
		&us.UserChatID,
	)
	if err != nil {
		logger.Log.Errorw("failed to get user state by chat id",
			"err", err,
		)
		return nil, err
	}

	return &us, nil
}

func (r *UserStateRepo) GetByThreadID(
	ctx context.Context,
	threadID int64,
) (*domain.UserState, error) {
	const q = `
			SELECT
				us.order_id,
				u.chat_id,
				o.status,
				e.id
			FROM users u
			JOIN user_state us
    		ON us.user_id = u.id
			LEFT JOIN orders o
				ON o.id = us.order_id
			LEFT JOIN experts e
				ON e.id = o.expert_id
			WHERE o.thread_id = $1
		`

	var us domain.UserState

	err := r.pool.QueryRow(ctx, q, threadID).Scan(
		&us.OrderID,
		&us.UserChatID,
		&us.OrderStatus,
		&us.ExpertID,
	)
	if err != nil {
		logger.Log.Errorw("failed to get user state by thread id",
			"err", err,
		)
		return nil, err
	}

	return &us, nil
}
