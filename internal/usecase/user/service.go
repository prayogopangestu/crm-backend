package user

import (
	"context"
	"net/mail"
	"strings"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/repositories"
	"github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Name        string
	CompanyName string
	Email       string
	Password    string
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginResult struct {
	Token string
	User  entities.User
}

type UpdateProfileInput struct {
	FirstName string
	LastName  string
	Email     string
}

type InviteInput struct {
	Name  string
	Email string
	Role  string
}

type InviteResult struct {
	User      entities.User
	InviteURL string
}

type Service struct {
	repository repositories.UserRepository
	cache      support.CacheHelper
	tokens     domain.TokenManager
	baseURL    string
	bcryptCost int
}

func NewService(repository repositories.UserRepository, cache support.CacheHelper, tokens domain.TokenManager, baseURL string, bcryptCost int) *Service {
	return &Service{
		repository: repository, cache: cache, tokens: tokens,
		baseURL: strings.TrimRight(baseURL, "/"), bcryptCost: bcryptCost,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (entities.User, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.CompanyName = strings.TrimSpace(input.CompanyName)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if len(input.Name) < 2 || len(input.CompanyName) < 2 || len(input.Password) < 6 {
		return entities.User{}, domain.ErrInvalidInput
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		return entities.User{}, domain.ErrInvalidInput
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), s.bcryptCost)
	if err != nil {
		return entities.User{}, err
	}
	first, last := splitName(input.Name)
	return s.repository.Register(ctx, input.CompanyName, entities.User{
		FirstName: first, LastName: last, Email: input.Email,
		PasswordHash: string(hash), Role: domain.RoleAdmin,
	})
}

func (s *Service) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	value, err := s.repository.ByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		if err == domain.ErrNotFound {
			return LoginResult{}, domain.ErrUnauthorized
		}
		return LoginResult{}, err
	}
	if value.Status != "Aktif" || value.PasswordHash == "" {
		return LoginResult{}, domain.ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(value.PasswordHash), []byte(input.Password)); err != nil {
		return LoginResult{}, domain.ErrUnauthorized
	}
	token, err := s.tokens.Create(value.ID, value.OrganizationID, value.Role, value.Name)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{Token: token, User: entities.User{ID: value.ID, Name: value.Name, Role: value.Role}}, nil
}

func (s *Service) AcceptInvite(ctx context.Context, token, password string) (entities.User, error) {
	if token == "" || len(password) < 6 {
		return entities.User{}, domain.ErrInvalidInput
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return entities.User{}, err
	}
	return s.repository.AcceptInvite(ctx, tokenHash(token), string(hash))
}

func (s *Service) Profile(ctx context.Context, principal domain.Principal) (entities.User, error) {
	var value entities.User
	key := "crm:" + principal.OrganizationID + ":profile:" + principal.UserID
	err := s.cache.Load(ctx, key, profileTTL, &value, func() error {
		var err error
		value, err = s.repository.ByID(ctx, principal.OrganizationID, principal.UserID)
		return err
	})
	return value, err
}

func (s *Service) UpdateProfile(ctx context.Context, principal domain.Principal, input UpdateProfileInput) (entities.User, error) {
	if strings.TrimSpace(input.FirstName) == "" || strings.TrimSpace(input.LastName) == "" {
		return entities.User{}, domain.ErrInvalidInput
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		return entities.User{}, domain.ErrInvalidInput
	}
	value, err := s.repository.UpdateProfile(ctx, principal, input.FirstName, input.LastName, input.Email)
	if err == nil {
		s.cache.InvalidateProfile(ctx, principal.OrganizationID, principal.UserID)
	}
	return value, err
}

func (s *Service) ListTeam(ctx context.Context, principal domain.Principal) ([]entities.User, error) {
	if err := domain.RequireAdmin(principal); err != nil {
		return nil, err
	}
	return s.repository.ListTeam(ctx, principal.OrganizationID)
}

func (s *Service) InviteMember(ctx context.Context, principal domain.Principal, input InviteInput) (InviteResult, error) {
	if err := domain.RequireAdmin(principal); err != nil {
		return InviteResult{}, err
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if len(input.Name) < 2 || (input.Role != domain.RoleAdmin && input.Role != domain.RoleSales) {
		return InviteResult{}, domain.ErrInvalidInput
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		return InviteResult{}, domain.ErrInvalidInput
	}
	plain, hash, err := randomToken()
	if err != nil {
		return InviteResult{}, err
	}
	first, last := splitName(input.Name)
	value, err := s.repository.InviteMember(ctx, principal, entities.User{
		FirstName: first, LastName: last, Email: input.Email, Role: input.Role,
	}, entities.Invitation{TokenHash: hash, ExpiresAt: inviteExpiry()})
	if err != nil {
		return InviteResult{}, err
	}
	return InviteResult{User: value, InviteURL: s.baseURL + "/accept-invite?token=" + plain}, nil
}

func (s *Service) RevokeMember(ctx context.Context, principal domain.Principal, userID string) error {
	if err := domain.RequireAdmin(principal); err != nil {
		return err
	}
	if userID == principal.UserID {
		return domain.ErrInvalidInput
	}
	err := s.repository.RevokeMember(ctx, principal.OrganizationID, userID)
	if err == nil {
		s.cache.InvalidateProfile(ctx, principal.OrganizationID, userID)
	}
	return err
}
