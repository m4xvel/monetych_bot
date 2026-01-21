package usecase

import (
	"context"
	"sync"

	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

type SupportService struct {
	repo    domain.SupportRepository
	support domain.Support
	mu      sync.RWMutex
}

func NewSupportService(r domain.SupportRepository) *SupportService {
	return &SupportService{repo: r}
}

func (s *SupportService) InitCache(ctx context.Context) error {
	row, err := s.repo.Get(ctx)
	if err != nil {
		logger.Log.Errorw("failed to load support config",
			"err", err,
		)
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.support = domain.Support{
		ID:       row.ID,
		ChatID:   row.ChatID,
		ChatLink: row.ChatLink,
	}

	logger.Log.Infow("support config loaded",
		"support_id", s.support.ID,
		"chat_id", s.support.ChatID,
	)

	return nil
}

func (s *SupportService) GetSupport() domain.Support {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.support
}
