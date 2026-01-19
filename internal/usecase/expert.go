package usecase

import (
	"context"
	"sync"

	"github.com/m4xvel/monetych_bot/internal/domain"
)

type ExpertService struct {
	repo    domain.ExpertRepository
	experts map[int]domain.Expert
	mu      sync.RWMutex
}

func NewExpertService(r domain.ExpertRepository) *ExpertService {
	return &ExpertService{
		repo:    r,
		experts: make(map[int]domain.Expert),
	}
}

func (s *ExpertService) InitCache(ctx context.Context) error {
	rows, err := s.repo.Get(ctx)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, r := range rows {
		if _, ok := s.experts[r.ID]; !ok {
			s.experts[int(r.ID)] = domain.Expert{
				ID:      r.ID,
				ChatID:  r.ChatID,
				TopicID: r.TopicID,
			}
		}
	}

	return nil
}

func (s *ExpertService) GetAllExperts() ([]domain.Expert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	experts := make([]domain.Expert, 0, len(s.experts))
	for _, e := range s.experts {
		experts = append(experts, e)
	}

	return experts, nil
}

func (s *ExpertService) GetExpertByID(id int) (domain.Expert, error) {
	s.mu.RLock()
	e, ok := s.experts[id]
	s.mu.RUnlock()

	if !ok {
		return domain.Expert{}, ErrNotFound
	}
	return e, nil
}
