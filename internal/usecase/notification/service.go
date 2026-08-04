package notification

import (
	"context"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/repositories"
)

type Service struct {
	repository repositories.NotificationRepository
}

func NewService(repository repositories.NotificationRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, principal domain.Principal) ([]entities.Notification, error) {
	if err := domain.RequireCanReadCRM(principal); err != nil {
		return nil, err
	}
	return s.repository.List(ctx, principal)
}

func (s *Service) Read(ctx context.Context, principal domain.Principal, id string) error {
	if err := domain.RequireCanReadCRM(principal); err != nil {
		return err
	}
	return s.repository.Read(ctx, principal, id)
}

func (s *Service) ReadAll(ctx context.Context, principal domain.Principal) error {
	if err := domain.RequireCanReadCRM(principal); err != nil {
		return err
	}
	return s.repository.ReadAll(ctx, principal)
}
