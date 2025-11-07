package usecase

import (
	"context"
	"fmt"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type GameService struct {
	repo domain.GameRepository
}

func NewGameService(repo domain.GameRepository) *GameService {
	return &GameService{repo: repo}
}

func (s *GameService) ListGames(ctx context.Context) ([]domain.Game, error) {
	games, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list games: %w", err)
	}
	return games, nil
}

func (s *GameService) ListGameTypes(
	ctx context.Context, gameID int) ([]string, error) {
	types, err := s.repo.GetGameTypeByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to list game types: %w", err)
	}
	return types, nil
}
