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
	rows, err := r.pool.Query(ctx, "SELECT id, name from games")
	if err != nil {
		return nil, fmt.Errorf("failed to query games: %w", err)
	}
	defer rows.Close()

	var games []domain.Game
	for rows.Next() {
		var g domain.Game
		if err := rows.Scan(&g.ID, &g.Name); err != nil {
			return nil, fmt.Errorf("failed to scan game row: %w", err)
		}
		games = append(games, g)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return games, nil
}

func (r *GameRepo) GetGameTypeByID(
	ctx context.Context,
	gameID int) ([]string, error) {
	rows, err := r.pool.Query(ctx, "SELECT gt.name FROM game_types gt JOIN game_type_links gtl ON gtl.type_id = gt.id JOIN games g ON g.id = gtl.game_id WHERE g.id = $1", gameID)
	if err != nil {
		return nil, fmt.Errorf("Failed to request game types: %w", err)
	}
	defer rows.Close()
	var types []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("failed to scan game types row: %w", err)
		}
		types = append(types, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return types, nil
}
