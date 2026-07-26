package deal

import (
	"context"
	"strings"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/repositories"
	"github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
)

type Input struct {
	Title      string
	Company    string
	Value      int64
	Priority   string
	Stage      string
	AssigneeID string
	LostReason string
}

type StageInput struct {
	Stage      string
	LostReason string
}

type Service struct {
	repository repositories.DealRepository
	cache      support.CacheHelper
}

func NewService(repository repositories.DealRepository, cache support.CacheHelper) *Service {
	return &Service{repository: repository, cache: cache}
}

func (s *Service) List(ctx context.Context, principal domain.Principal) ([]entities.Deal, error) {
	return s.repository.List(ctx, principal.OrganizationID)
}

func (s *Service) Create(ctx context.Context, principal domain.Principal, input Input) (entities.Deal, error) {
	if err := validate(input); err != nil {
		return entities.Deal{}, err
	}
	result, err := s.repository.Create(ctx, principal, toEntity(input, principal))
	if err == nil {
		s.cache.InvalidateCRM(ctx, principal.OrganizationID)
	}
	return result, err
}

func (s *Service) Update(ctx context.Context, principal domain.Principal, id string, input Input) (entities.Deal, error) {
	if err := validate(input); err != nil {
		return entities.Deal{}, err
	}
	result, err := s.repository.Update(ctx, principal, id, toEntity(input, principal))
	if err == nil {
		s.cache.InvalidateCRM(ctx, principal.OrganizationID)
	}
	return result, err
}

func (s *Service) UpdateStage(ctx context.Context, principal domain.Principal, id string, input StageInput) error {
	if input.Stage == "" || (input.Stage == "lost" && strings.TrimSpace(input.LostReason) == "") {
		return domain.ErrInvalidInput
	}
	if input.Stage != "lost" {
		input.LostReason = ""
	}
	err := s.repository.UpdateStage(ctx, principal, id, input.Stage, input.LostReason)
	if err == nil {
		s.cache.InvalidateCRM(ctx, principal.OrganizationID)
	}
	return err
}

func (s *Service) Delete(ctx context.Context, principal domain.Principal, id string) error {
	err := s.repository.Delete(ctx, principal, id)
	if err == nil {
		s.cache.InvalidateCRM(ctx, principal.OrganizationID)
	}
	return err
}

func validate(input Input) error {
	if len(strings.TrimSpace(input.Title)) < 2 || len(strings.TrimSpace(input.Company)) < 2 ||
		input.Value < 0 || input.Stage == "" {
		return domain.ErrInvalidInput
	}
	if input.Priority != "High" && input.Priority != "Medium" && input.Priority != "Low" {
		return domain.ErrInvalidInput
	}
	if input.Stage == "lost" && strings.TrimSpace(input.LostReason) == "" {
		return domain.ErrInvalidInput
	}
	return nil
}

func toEntity(input Input, principal domain.Principal) entities.Deal {
	return entities.Deal{
		Title: input.Title, Company: input.Company, Value: input.Value,
		Priority: input.Priority, Stage: input.Stage, LostReason: input.LostReason,
		Assignee: entities.DealAssignee{ID: input.AssigneeID},
	}
}
