package task

import (
	"context"
	"strings"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/repositories"
	"github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
)

type Input struct {
	Title      string
	Company    string
	Time       string
	Date       string
	Type       string
	Priority   string
	Assignee   string
	AssigneeID string
	Notes      string
	Completed  bool
}

type Service struct {
	repository repositories.TaskRepository
	cache      support.CacheHelper
	location   *time.Location
}

func NewService(repository repositories.TaskRepository, cache support.CacheHelper, location *time.Location) *Service {
	return &Service{repository: repository, cache: cache, location: location}
}

func (s *Service) List(ctx context.Context, principal domain.Principal, date, status string) ([]entities.Task, error) {
	if err := domain.RequireCanReadCRM(principal); err != nil {
		return nil, err
	}
	if date != "" {
		if _, err := time.ParseInLocation("2006-01-02", date, s.location); err != nil {
			return nil, domain.ErrInvalidInput
		}
	}
	if status != "" && status != "overdue" && status != "today" && status != "upcoming" {
		return nil, domain.ErrInvalidInput
	}
	return s.repository.List(ctx, principal.OrganizationID, date, status, s.location)
}

func (s *Service) Create(ctx context.Context, principal domain.Principal, input Input) (entities.Task, error) {
	if err := domain.RequireCanWriteCRM(principal); err != nil {
		return entities.Task{}, err
	}
	if err := validate(input, false); err != nil {
		return entities.Task{}, err
	}
	result, err := s.repository.Create(ctx, principal, toEntity(input))
	if err == nil {
		s.cache.InvalidateCRM(ctx, principal.OrganizationID)
	}
	return result, err
}

func (s *Service) Update(ctx context.Context, principal domain.Principal, id string, input Input) (entities.Task, error) {
	if err := domain.RequireCanWriteCRM(principal); err != nil {
		return entities.Task{}, err
	}
	if err := validate(input, true); err != nil {
		return entities.Task{}, err
	}
	result, err := s.repository.Update(ctx, principal, id, toEntity(input))
	if err == nil {
		s.cache.InvalidateCRM(ctx, principal.OrganizationID)
	}
	return result, err
}

func (s *Service) Toggle(ctx context.Context, principal domain.Principal, id string, completed bool) error {
	if err := domain.RequireCanWriteCRM(principal); err != nil {
		return err
	}
	err := s.repository.Toggle(ctx, principal, id, completed)
	if err == nil {
		s.cache.InvalidateCRM(ctx, principal.OrganizationID)
	}
	return err
}

func (s *Service) Delete(ctx context.Context, principal domain.Principal, id string) error {
	if err := domain.RequireCanWriteCRM(principal); err != nil {
		return err
	}
	err := s.repository.Delete(ctx, principal, id)
	if err == nil {
		s.cache.InvalidateCRM(ctx, principal.OrganizationID)
	}
	return err
}

func validate(input Input, partial bool) error {
	if !partial && (len(strings.TrimSpace(input.Title)) < 3 || len(strings.TrimSpace(input.Company)) < 2 ||
		input.Date == "" || input.Time == "" || input.Type == "" || input.Priority == "") {
		return domain.ErrInvalidInput
	}
	if input.Date != "" {
		if _, err := time.Parse("2006-01-02", input.Date); err != nil {
			return domain.ErrInvalidInput
		}
	}
	if input.Time != "" {
		if _, err := time.Parse("15:04", input.Time); err != nil {
			return domain.ErrInvalidInput
		}
	}
	if input.Type != "" && input.Type != "Meeting" && input.Type != "Call" && input.Type != "Proposal" && input.Type != "Other" {
		return domain.ErrInvalidInput
	}
	if input.Priority != "" && input.Priority != "Tinggi" && input.Priority != "Sedang" && input.Priority != "Rendah" {
		return domain.ErrInvalidInput
	}
	return nil
}

func toEntity(input Input) entities.Task {
	return entities.Task{
		Title: input.Title, Company: input.Company, Date: input.Date, Time: input.Time,
		Type: input.Type, Priority: input.Priority, Notes: input.Notes, Completed: input.Completed,
		Assignee: input.Assignee, AssigneeID: input.AssigneeID,
	}
}
