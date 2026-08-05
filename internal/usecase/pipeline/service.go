package pipeline

import (
	"context"
	"strings"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/repositories"
	"github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
)

type Input struct {
	Name  string
	Color string
}

type Service struct {
	repository repositories.PipelineRepository
	cache      support.CacheHelper
}

func NewService(repository repositories.PipelineRepository, cache support.CacheHelper) *Service {
	return &Service{repository: repository, cache: cache}
}

func (s *Service) List(ctx context.Context, principal domain.Principal) ([]entities.Stage, error) {
	if err := domain.RequireCanReadCRM(principal); err != nil {
		return nil, err
	}
	return s.repository.List(ctx, principal.OrganizationID)
}

func (s *Service) Create(ctx context.Context, principal domain.Principal, input Input) (entities.Stage, error) {
	if err := domain.RequireAdmin(principal); err != nil {
		return entities.Stage{}, err
	}
	if len(strings.TrimSpace(input.Name)) < 2 {
		return entities.Stage{}, domain.ErrInvalidInput
	}
	result, err := s.repository.Create(ctx, principal.OrganizationID, entities.Stage{Name: input.Name, Color: input.Color})
	if err == nil {
		s.cache.InvalidateCRM(ctx, principal.OrganizationID)
	}
	return result, err
}

func (s *Service) Reorder(ctx context.Context, principal domain.Principal, ids []string) error {
	if err := domain.RequireAdmin(principal); err != nil {
		return err
	}
	if len(ids) == 0 {
		return domain.ErrInvalidInput
	}
	return s.repository.Reorder(ctx, principal.OrganizationID, ids)
}

func (s *Service) Delete(ctx context.Context, principal domain.Principal, id string) error {
	if err := domain.RequireAdmin(principal); err != nil {
		return err
	}
	return s.repository.Delete(ctx, principal.OrganizationID, id)
}
