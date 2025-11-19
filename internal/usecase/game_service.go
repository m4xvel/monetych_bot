package usecase

import (
	"context"
	"fmt"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type GameService struct {
	repo domain.GameRepository
}

func NewGameService(r domain.GameRepository) *GameService {
	return &GameService{repo: r}
}

func (s *GameService) ListGames(ctx context.Context) ([]domain.Game, error) {
	gs, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list games: %w", err)
	}
	return gs, nil
}

func (s *GameService) ListGameTypes(ctx context.Context, gameID int) ([]string, error) {
	t, err := s.repo.GetTypes(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("list game types: %w", err)
	}
	return t, nil
}
