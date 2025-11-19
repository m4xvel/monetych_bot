package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

type GameRepo struct {
	pool *pgxpool.Pool
}

func NewGameRepo(pool *pgxpool.Pool) *GameRepo {
	return &GameRepo{pool: pool}
}

func (r *GameRepo) GetAll(ctx context.Context) ([]domain.Game, error) {
	const q = `SELECT id, name FROM games`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("game repo - query GetAll: %w", err)
	}
	defer rows.Close()

	var out []domain.Game
	for rows.Next() {
		var g domain.Game
		if err := rows.Scan(&g.ID, &g.Name); err != nil {
			return nil, fmt.Errorf("game repo - scan GetAll: %w", err)
		}
		out = append(out, g)
	}
	return out, nil
}

func (r *GameRepo) GetTypes(ctx context.Context, gameID int) ([]string, error) {
	const q = `
	SELECT gt.name
	FROM game_types gt
	JOIN game_type_links gtl ON gtl.type_id = gt.id
	WHERE gtl.game_id = $1`
	rows, err := r.pool.Query(ctx, q, gameID)
	if err != nil {
		return nil, fmt.Errorf("game repo - query GetTypes: %w", err)
	}
	defer rows.Close()

	var types []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("game repo - scan GetTypes: %w", err)
		}
		types = append(types, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("game repo - rows err GetTypes: %w", err)
	}
	return types, nil
}
