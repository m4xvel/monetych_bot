package postgres

import (
	"context"
	"fmt"

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
	titles []string,
) error {
	if len(titles) == 0 {
		return fmt.Errorf("titles list is empty")
	}

	const qRequiredCount = `
		SELECT COUNT(*)
		FROM (
			SELECT DISTINCT ON (p.title) p.id
			FROM policies p
			WHERE p.title = ANY($1::text[])
			ORDER BY p.title, p.created_at DESC, p.id DESC
		) latest;
	`

	var requiredCount int
	if err := r.pool.QueryRow(ctx, qRequiredCount, titles).Scan(&requiredCount); err != nil {
		return dbErr("user_policy_acceptances.required_count", err)
	}

	if requiredCount != len(titles) {
		return fmt.Errorf("required policies not found: expected=%d got=%d", len(titles), requiredCount)
	}

	const qInsert = `
		INSERT INTO user_policy_acceptances (user_id, policy_id)
		SELECT u.id, latest.id
		FROM users u
		JOIN (
			SELECT DISTINCT ON (p.title) p.id
			FROM policies p
			WHERE p.title = ANY($2::text[])
			ORDER BY p.title, p.created_at DESC, p.id DESC
		) latest ON TRUE
		WHERE u.chat_id = $1
		ON CONFLICT DO NOTHING;
	`

	_, err := r.pool.Exec(ctx, qInsert, chatID, titles)
	if err != nil {
		return dbErr("user_policy_acceptances.set", err)
	}
	return nil
}

func (r *UserPolicyAcceptancesRepo) IsUserAccepted(
	ctx context.Context,
	chatID int64,
	titles []string,
) (bool, error) {
	if len(titles) == 0 {
		return false, nil
	}

	const q = `
		WITH
		required AS (
			SELECT DISTINCT ON (p.title) p.id
			FROM policies p
			WHERE p.title = ANY($2::text[])
			ORDER BY p.title, p.created_at DESC, p.id DESC
		),
		user_row AS (
			SELECT u.id
			FROM users u
			WHERE u.chat_id = $1
		),
		accepted AS (
			SELECT upa.policy_id
			FROM user_policy_acceptances upa
			JOIN user_row ur ON ur.id = upa.user_id
		)
		SELECT
			(SELECT COUNT(*) FROM required) = cardinality($2::text[])
			AND EXISTS (SELECT 1 FROM user_row)
			AND COUNT(a.policy_id) = COUNT(*)
		FROM required r
		LEFT JOIN accepted a ON a.policy_id = r.id
	`

	var exists bool

	err := r.pool.QueryRow(ctx, q, chatID, titles).Scan(&exists)
	if err != nil {
		return false, dbErr("user_policy_acceptances.is_accepted", err)
	}
	return exists, nil
}
