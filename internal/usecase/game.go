package usecase

import (
	"context"
	"sync"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type GameService struct {
	repo domain.GameRepository

	games       map[int]domain.Game
	types       map[int]domain.GameType
	gameToTypes map[int][]int

	mu sync.RWMutex
}

func NewGameService(r domain.GameRepository) *GameService {
	return &GameService{
		repo:        r,
		games:       make(map[int]domain.Game),
		types:       make(map[int]domain.GameType),
		gameToTypes: make(map[int][]int),
	}
}

func (s *GameService) InitCache(ctx context.Context) error {
	rows, err := s.repo.Get(ctx)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, r := range rows {
		if _, ok := s.games[r.GameID]; !ok {
			s.games[r.GameID] = domain.Game{
				ID:   r.GameID,
				Name: r.GameName,
			}
			s.gameToTypes[r.GameID] = []int{}
		}

		if r.TypeID != nil {
			if _, ok := s.types[*r.TypeID]; !ok {
				s.types[*r.TypeID] = domain.GameType{
					ID:   *r.TypeID,
					Name: *r.TypeName,
				}
			}
			s.gameToTypes[r.GameID] =
				append(s.gameToTypes[r.GameID], *r.TypeID)
		}
	}

	return nil
}

func (s *GameService) GetGameByID(id int) (domain.Game, error) {
	s.mu.RLock()
	g, ok := s.games[id]
	s.mu.RUnlock()

	if !ok {
		return domain.Game{}, ErrNotFound
	}

	return g, nil
}

func (s *GameService) GetTypeByID(id int) (domain.GameType, error) {
	s.mu.RLock()
	t, ok := s.types[id]
	s.mu.RUnlock()

	if !ok {
		return domain.GameType{}, ErrNotFound
	}

	return t, nil
}

func (s *GameService) GetAllGames() ([]domain.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	games := make([]domain.Game, 0, len(s.games))
	for _, g := range s.games {
		games = append(games, g)
	}

	return games, nil
}

func (s *GameService) GetGameTypesByGameID(
	id int,
) ([]domain.GameType, error) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	typeIDs, ok := s.gameToTypes[id]
	if !ok {
		return nil, ErrNotFound
	}

	out := make([]domain.GameType, 0, len(typeIDs))
	for _, id := range typeIDs {
		out = append(out, s.types[id])
	}

	return out, nil
}
