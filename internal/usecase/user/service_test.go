package user

import (
	"context"
	"testing"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
)

type userRepoStub struct {
	membershipErr error
	workspaces    []entities.WorkspaceSummary
}

func (r userRepoStub) Register(context.Context, string, entities.User) (entities.User, error) {
	return entities.User{}, nil
}
func (r userRepoStub) CreateGoogleUser(context.Context, string, entities.User) (entities.User, error) {
	return entities.User{}, nil
}
func (r userRepoStub) ByEmail(context.Context, string) (entities.User, error) {
	return entities.User{}, nil
}
func (r userRepoStub) ByID(context.Context, string, string) (entities.User, error) {
	return entities.User{}, nil
}
func (r userRepoStub) UpdateProfile(context.Context, domain.Principal, string, string, string) (entities.User, error) {
	return entities.User{}, nil
}
func (r userRepoStub) AcceptInvite(context.Context, string, string) (entities.User, error) {
	return entities.User{}, nil
}
func (r userRepoStub) LinkGoogleID(context.Context, string, string, string, string) error { return nil }
func (r userRepoStub) ListTeam(context.Context, string) ([]entities.User, error)          { return nil, nil }
func (r userRepoStub) InviteMember(context.Context, domain.Principal, entities.User, entities.Invitation) (entities.User, error) {
	return entities.User{}, nil
}
func (r userRepoStub) RevokeMember(context.Context, string, string) error { return nil }
func (r userRepoStub) ListWorkspaces(context.Context, string) ([]entities.WorkspaceSummary, error) {
	return r.workspaces, nil
}
func (r userRepoStub) Membership(context.Context, string, string) (entities.OrganizationMember, error) {
	if r.membershipErr != nil {
		return entities.OrganizationMember{}, r.membershipErr
	}
	return entities.OrganizationMember{Role: domain.RoleViewer}, nil
}
func (r userRepoStub) ByEmailGlobal(context.Context, string) (entities.User, error) {
	return entities.User{}, nil
}
func (r userRepoStub) CreatePersonalWorkspace(context.Context, string, entities.User) (entities.OrganizationMember, error) {
	return entities.OrganizationMember{}, nil
}
func (r userRepoStub) AddMembership(context.Context, string, string, string, *string) (entities.OrganizationMember, error) {
	return entities.OrganizationMember{}, nil
}
func (r userRepoStub) AcceptInviteForUser(context.Context, string, string) (entities.User, error) {
	return entities.User{}, nil
}
func (r userRepoStub) LookupInvitation(context.Context, string) (entities.Invitation, error) {
	return entities.Invitation{}, nil
}
func (r userRepoStub) MarkInvitationAccepted(context.Context, string, string) error { return nil }
func (r userRepoStub) ActiveMembershipExists(context.Context, string, string) (bool, error) {
	return false, nil
}

type tokenStub struct{}

func (tokenStub) Create(string, string, string, string) (string, error) { return "token", nil }
func (tokenStub) Parse(string) (domain.Principal, error)                { return domain.Principal{}, nil }

func TestSwitchWorkspaceRejectsNonMember(t *testing.T) {
	service := NewService(userRepoStub{membershipErr: domain.ErrNotFound}, support.CacheHelper{}, tokenStub{}, "", 4)
	_, err := service.SwitchWorkspace(context.Background(),
		domain.Principal{UserID: "user-1", OrganizationID: "org-1", Role: domain.RoleOwner, Name: "Sarah"},
		SwitchWorkspaceInput{WorkspaceID: "org-2"},
	)
	if err != domain.ErrForbidden {
		t.Fatalf("expected forbidden for non-member switch, got %v", err)
	}
}

func TestSwitchWorkspaceUsesMembershipRole(t *testing.T) {
	service := NewService(userRepoStub{
		workspaces: []entities.WorkspaceSummary{{ID: "org-2", Name: "Shared CRM", Role: domain.RoleViewer}},
	}, support.CacheHelper{}, tokenStub{}, "", 4)
	result, err := service.SwitchWorkspace(context.Background(),
		domain.Principal{UserID: "user-1", OrganizationID: "org-1", Role: domain.RoleOwner, Name: "Sarah"},
		SwitchWorkspaceInput{WorkspaceID: "org-2"},
	)
	if err != nil {
		t.Fatalf("switch should succeed: %v", err)
	}
	if result.Workspace.Role != domain.RoleViewer {
		t.Fatalf("expected viewer role from membership, got %s", result.Workspace.Role)
	}
}
