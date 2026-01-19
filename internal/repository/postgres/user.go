package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
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
		return ErrAdd
	}

	if tag.RowsAffected() == 0 {
		return ErrUserAlreadyExists
	}

	return nil
}

func (r *UserRepo) UpdatePhoto(ctx context.Context, user domain.User) error {
	const q = `
	UPDATE users
	SET img_url = $1
	WHERE chat_id = $2
	`
	_, err := r.pool.Exec(ctx, q, user.PhotoURL, user.ChatID)
	return err
}

func (r *UserRepo) Get(ctx context.Context, user domain.User) (*domain.User, error) {
	const q = `
	SELECT id, chat_id, name
	FROM users
	WHERE chat_id = $1
	`
	var u domain.User

	err := r.pool.QueryRow(ctx, q, user.ChatID).Scan(&u.ID, &u.ChatID, &u.Name)
	if err != nil {
		return nil, err
	}

	return &u, nil
}
