package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

type GameRepo struct {
	pool *pgxpool.Pool
}

func NewGameRepo(pool *pgxpool.Pool) *GameRepo {
	return &GameRepo{pool: pool}
}

func (r *GameRepo) Get(
	ctx context.Context,
) ([]domain.GameWithTypeRow, error) {
	const q = `
	SELECT 
		g.id,
		g.name,
		gt.id,
		gt.name
	FROM games g
	LEFT JOIN game_type_links gtl ON gtl.game_id = g.id
	LEFT JOIN game_types gt ON gt.id = gtl.game_type_id
	ORDER BY g.id
	`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.GameWithTypeRow

	for rows.Next() {
		var r domain.GameWithTypeRow
		if err := rows.Scan(
			&r.GameID,
			&r.GameName,
			&r.TypeID,
			&r.TypeName,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
