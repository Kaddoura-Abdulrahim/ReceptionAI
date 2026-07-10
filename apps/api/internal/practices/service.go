package practices

import (
	"context"
	"strings"

	"dentaldesk/apps/api/internal/domain"
	"dentaldesk/apps/api/internal/store"
)

type Service struct {
	store store.Store
}

func NewService(store store.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Create(ctx context.Context, userID, name string) domain.Practice {
	_ = ctx
	return s.store.CreatePractice(strings.TrimSpace(name), userID)
}

func (s *Service) ListForUser(ctx context.Context, userID string) []domain.Practice {
	_ = ctx
	return s.store.ListPracticesForUser(userID)
}

func (s *Service) IsMember(ctx context.Context, practiceID, userID string) bool {
	_ = ctx
	return s.store.IsMember(practiceID, userID)
}
