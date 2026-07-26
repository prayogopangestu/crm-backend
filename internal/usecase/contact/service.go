package contact

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/repositories"
	"github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
)

var statuses = map[string]bool{
	"Negosiasi": true, "Menang": true, "Prospek Awal": true,
	"Proposal": true, "Kalah": true, "Kualifikasi": true,
}

type Input struct {
	Name      string
	Email     string
	Company   string
	Role      string
	Status    string
	AvatarURL string
}

type Service struct {
	repository repositories.ContactRepository
	cache      support.CacheHelper
}

func NewService(repository repositories.ContactRepository, cache support.CacheHelper) *Service {
	return &Service{repository: repository, cache: cache}
}

func (s *Service) List(ctx context.Context, principal domain.Principal, search, status string, page, limit int) (entities.ContactList, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if status != "" && !statuses[status] {
		return entities.ContactList{}, domain.ErrInvalidInput
	}
	return s.repository.List(ctx, principal.OrganizationID, search, status, entities.ContactPage{Page: page, Limit: limit})
}

func (s *Service) Create(ctx context.Context, principal domain.Principal, input Input) (entities.Contact, error) {
	if err := validate(input); err != nil {
		return entities.Contact{}, err
	}
	result, err := s.repository.Create(ctx, principal, toEntity(input))
	if err == nil {
		s.cache.InvalidateCRM(ctx, principal.OrganizationID)
	}
	return result, err
}

func (s *Service) Update(ctx context.Context, principal domain.Principal, id string, input Input) (entities.Contact, error) {
	if err := validate(input); err != nil {
		return entities.Contact{}, err
	}
	result, err := s.repository.Update(ctx, principal, id, toEntity(input))
	if err == nil {
		s.cache.InvalidateCRM(ctx, principal.OrganizationID)
	}
	return result, err
}

func (s *Service) Delete(ctx context.Context, principal domain.Principal, id string) error {
	err := s.repository.Delete(ctx, principal, id)
	if err == nil {
		s.cache.InvalidateCRM(ctx, principal.OrganizationID)
	}
	return err
}

func validate(input Input) error {
	if len(strings.TrimSpace(input.Name)) < 2 || len(strings.TrimSpace(input.Company)) < 2 || !statuses[input.Status] {
		return domain.ErrInvalidInput
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		return domain.ErrInvalidInput
	}
	return nil
}

func toEntity(input Input) entities.Contact {
	return entities.Contact{
		Name: input.Name, Email: input.Email, Company: input.Company,
		Role: input.Role, Status: input.Status, AvatarURL: input.AvatarURL, LastContactedAt: time.Now(),
	}
}
