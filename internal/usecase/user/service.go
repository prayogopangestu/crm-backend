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

type GoogleProfileInput struct {
	GoogleID  string
	Email     string
	FirstName string
	LastName  string
	AvatarURL string
}

type LoginResult struct {
	Token      string
	User       entities.User
	Workspaces []entities.WorkspaceSummary
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

// SwitchWorkspaceInput selects the active workspace for the authenticated user.
type SwitchWorkspaceInput struct {
	WorkspaceID string
}

// SwitchWorkspaceResult is returned after switching the active workspace: a
// fresh token scoped to the selected workspace plus its summary.
type SwitchWorkspaceResult struct {
	Token     string
	Workspace entities.WorkspaceSummary
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
		PasswordHash: string(hash), Role: domain.RoleOwner,
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
	active, workspaces, err := s.resolveActiveWorkspace(ctx, value.ID)
	if err != nil {
		return LoginResult{}, err
	}
	token, err := s.tokens.Create(value.ID, active.ID, active.Role, value.Name)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
		Token: token, Workspaces: workspaces,
		User: entities.User{ID: value.ID, Name: value.Name, Role: active.Role},
	}, nil
}

// resolveActiveWorkspace returns the user's default/first active workspace and
// the full workspace list. A user without any active membership cannot log in.
func (s *Service) resolveActiveWorkspace(ctx context.Context, userID string) (entities.WorkspaceSummary, []entities.WorkspaceSummary, error) {
	workspaces, err := s.repository.ListWorkspaces(ctx, userID)
	if err != nil {
		return entities.WorkspaceSummary{}, nil, err
	}
	if len(workspaces) == 0 {
		return entities.WorkspaceSummary{}, nil, domain.ErrUnauthorized
	}
	return workspaces[0], workspaces, nil
}

func (s *Service) LoginWithGoogle(ctx context.Context, in GoogleProfileInput) (LoginResult, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if in.GoogleID == "" || in.Email == "" {
		return LoginResult{}, domain.ErrInvalidInput
	}

	existing, err := s.repository.ByEmail(ctx, in.Email)
	switch {
	case err == nil:
		if existing.Status != "Aktif" {
			return LoginResult{}, domain.ErrUnauthorized
		}
		if existing.GoogleID == nil || *existing.GoogleID == "" {
			if err := s.repository.LinkGoogleID(ctx, existing.OrganizationID, existing.ID, in.GoogleID, in.AvatarURL); err != nil {
				return LoginResult{}, err
			}
		}
		active, workspaces, err := s.resolveActiveWorkspace(ctx, existing.ID)
		if err != nil {
			return LoginResult{}, err
		}
		token, err := s.tokens.Create(existing.ID, active.ID, active.Role, existing.Name)
		if err != nil {
			return LoginResult{}, err
		}
		return LoginResult{
			Token: token, Workspaces: workspaces,
			User: entities.User{ID: existing.ID, Name: existing.Name, Role: active.Role},
		}, nil

	case err == domain.ErrNotFound:
		googleID := in.GoogleID
		created, err := s.repository.CreateGoogleUser(ctx, "Organisasi "+firstNonEmpty(in.FirstName, in.Email), entities.User{
			FirstName: in.FirstName, LastName: in.LastName, Email: in.Email,
			GoogleID: &googleID, AvatarURL: in.AvatarURL, Role: domain.RoleOwner,
		})
		if err != nil {
			return LoginResult{}, err
		}
		active, workspaces, err := s.resolveActiveWorkspace(ctx, created.ID)
		if err != nil {
			return LoginResult{}, err
		}
		token, err := s.tokens.Create(created.ID, active.ID, active.Role, created.Name)
		if err != nil {
			return LoginResult{}, err
		}
		return LoginResult{
			Token: token, Workspaces: workspaces,
			User: entities.User{ID: created.ID, Name: created.Name, Role: active.Role},
		}, nil

	default:
		return LoginResult{}, err
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return "Google User"
}

func (s *Service) AcceptInvite(ctx context.Context, token, password string) (LoginResult, error) {
	if token == "" || len(password) < 6 {
		return LoginResult{}, domain.ErrInvalidInput
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return LoginResult{}, err
	}
	user, err := s.repository.AcceptInvite(ctx, tokenHash(token), string(hash))
	if err != nil {
		return LoginResult{}, err
	}
	tok, err := s.tokens.Create(user.ID, user.OrganizationID, user.Role, user.Name)
	if err != nil {
		return LoginResult{}, err
	}
	workspaces, werr := s.repository.ListWorkspaces(ctx, user.ID)
	if werr != nil {
		workspaces = nil
	}
	return LoginResult{Token: tok, User: user, Workspaces: workspaces}, nil
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
	if err := domain.RequireWorkspaceAdmin(principal); err != nil {
		return nil, err
	}
	return s.repository.ListTeam(ctx, principal.OrganizationID)
}

func (s *Service) InviteMember(ctx context.Context, principal domain.Principal, input InviteInput) (InviteResult, error) {
	if err := domain.RequireWorkspaceAdmin(principal); err != nil {
		return InviteResult{}, err
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if !domain.InvitableRole(input.Role) {
		return InviteResult{}, domain.ErrInvalidInput
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		return InviteResult{}, domain.ErrInvalidInput
	}
	exists, err := s.repository.ActiveMembershipExists(ctx, principal.OrganizationID, input.Email)
	if err != nil {
		return InviteResult{}, err
	}
	if exists {
		return InviteResult{}, domain.ErrConflict
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
	if err := domain.RequireWorkspaceAdmin(principal); err != nil {
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

// ListWorkspaces returns every workspace the authenticated user can access.
func (s *Service) ListWorkspaces(ctx context.Context, principal domain.Principal) ([]entities.WorkspaceSummary, error) {
	return s.repository.ListWorkspaces(ctx, principal.UserID)
}

// SwitchWorkspace mints a fresh token scoped to a workspace the user is an
// active member of. Non-members are rejected with ErrForbidden.
func (s *Service) SwitchWorkspace(ctx context.Context, principal domain.Principal, input SwitchWorkspaceInput) (SwitchWorkspaceResult, error) {
	input.WorkspaceID = strings.TrimSpace(input.WorkspaceID)
	if input.WorkspaceID == "" {
		return SwitchWorkspaceResult{}, domain.ErrInvalidInput
	}
	membership, err := s.repository.Membership(ctx, principal.UserID, input.WorkspaceID)
	if err != nil {
		if err == domain.ErrNotFound {
			return SwitchWorkspaceResult{}, domain.ErrForbidden
		}
		return SwitchWorkspaceResult{}, err
	}
	workspaces, err := s.repository.ListWorkspaces(ctx, principal.UserID)
	if err != nil {
		return SwitchWorkspaceResult{}, err
	}
	var summary entities.WorkspaceSummary
	for _, ws := range workspaces {
		if ws.ID == input.WorkspaceID {
			summary = ws
			break
		}
	}
	if summary.ID == "" {
		summary = entities.WorkspaceSummary{ID: input.WorkspaceID, Role: membership.Role}
	}
	token, err := s.tokens.Create(principal.UserID, input.WorkspaceID, membership.Role, principal.Name)
	if err != nil {
		return SwitchWorkspaceResult{}, err
	}
	return SwitchWorkspaceResult{Token: token, Workspace: summary}, nil
}
