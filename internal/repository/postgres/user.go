package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/apperr"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Add(ctx context.Context, user domain.User) error {
	const q = `
	INSERT INTO users (chat_id, name)
	VALUES ($1, $2)
	ON CONFLICT (chat_id) DO NOTHING
	`

	tag, err := r.pool.Exec(ctx, q, user.ChatID, user.Name)
	if err != nil {
		wrapped := dbErr("user.add", err)
		logger.Log.Errorw("failed to insert user",
			"err", wrapped,
		)
		return wrapped
	}

	if tag.RowsAffected() == 0 {
		return nil
	}

	logger.Log.Infow("user created",
		"chat_id", user.ChatID,
		"name", user.Name,
	)

	return nil
}

func (r *UserRepo) UpdatePhoto(ctx context.Context, user domain.User) error {
	const q = `
	UPDATE users
	SET img_url = $1
	WHERE chat_id = $2
	`
	_, err := r.pool.Exec(ctx, q, user.PhotoURL, user.ChatID)
	if err != nil {
		wrapped := dbErr("user.update_photo", err)
		logger.Log.Errorw("failed to update user photo",
			"err", wrapped,
		)
		return wrapped
	}

	return nil
}

func (r *UserRepo) UpdateVerified(
	ctx context.Context,
	chatID int64,
	isVerified bool,
) error {
	const q = `
	UPDATE users
	SET is_verified = $1
	WHERE chat_id = $2
	`

	cmd, err := r.pool.Exec(ctx, q, isVerified, chatID)
	if err != nil {
		wrapped := dbErr("user.update_verified", err)
		logger.Log.Errorw("failed to update user verification status",
			"err", wrapped,
		)
		return wrapped
	}

	if cmd.RowsAffected() == 0 {
		return dbErrKind("user.update_verified", apperr.KindNotFound, nil)
	}

	return nil
}

func (r *UserRepo) Get(ctx context.Context, user domain.User) (*domain.User, error) {
	const q = `
	SELECT id, chat_id, name, is_verified
	FROM users
	WHERE chat_id = $1
	`
	var u domain.User
	err := r.pool.QueryRow(ctx, q, user.ChatID).
		Scan(&u.ID, &u.ChatID, &u.Name, &u.IsVerified)

	if err != nil {
		wrapped := dbErr("user.get", err)
		logger.Log.Errorw("failed to get user",
			"err", wrapped,
		)
		return nil, wrapped
	}

	return &u, nil
}

func (r *UserRepo) IncrementOrders(ctx context.Context, chatID int64) error {
	const q = `
		UPDATE users 
		SET total_orders = total_orders + 1 
		WHERE chat_id = $1
	`

	cmd, err := r.pool.Exec(ctx, q, chatID)
	if err != nil {
		wrapped := dbErr("user.increment_orders", err)
		logger.Log.Errorw("failed to increment user orders",
			"err", wrapped,
		)
		return wrapped
	}

	if cmd.RowsAffected() == 0 {
		return dbErrKind("user.increment_orders", apperr.KindNotFound, nil)
	}

	return nil
}
