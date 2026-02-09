package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserPolicyAcceptancesRepo struct {
	pool *pgxpool.Pool
}

func NewUserPolicyAcceptancesRepo(
	pool *pgxpool.Pool,
) *UserPolicyAcceptancesRepo {
	return &UserPolicyAcceptancesRepo{
		pool: pool,
	}
}

func (r *UserPolicyAcceptancesRepo) Set(
	ctx context.Context,
	chatID int64,
	version string,
) error {
	const q = `
		INSERT INTO user_policy_acceptances (user_id, policy_id)
		SELECT u.id, p.id
		FROM users u, policies p
		WHERE u.chat_id = $1
			AND p.version = $2
		ON CONFLICT DO NOTHING;
	`

	_, err := r.pool.Exec(ctx, q, chatID, version)
	if err != nil {
		return dbErr("user_policy_acceptances.set", err)
	}
	return nil
}

func (r *UserPolicyAcceptancesRepo) IsUserAccepted(
	ctx context.Context,
	chatID int64,
	version string,
) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1
			FROM user_policy_acceptances upa
			JOIN users u ON u.id = upa.user_id
			JOIN policies p ON p.id = upa.policy_id
			WHERE u.chat_id = $1
				AND p.version = $2
		)
	`

	var exists bool

	err := r.pool.QueryRow(ctx, q, chatID, version).Scan(&exists)
	if err != nil {
		return false, dbErr("user_policy_acceptances.is_accepted", err)
	}
	return exists, nil
}
