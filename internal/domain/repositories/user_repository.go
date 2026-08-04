package repositories

import (
	"context"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
)

type UserRepository interface {
	Register(ctx context.Context, orgName string, user entities.User) (entities.User, error)
	CreateGoogleUser(ctx context.Context, orgName string, user entities.User) (entities.User, error)
	ByEmail(ctx context.Context, email string) (entities.User, error)
	ByID(ctx context.Context, organizationID, userID string) (entities.User, error)
	UpdateProfile(ctx context.Context, principal domain.Principal, firstName, lastName, email string) (entities.User, error)
	AcceptInvite(ctx context.Context, tokenHash, passwordHash string) (entities.User, error)
	LinkGoogleID(ctx context.Context, organizationID, userID, googleID, avatarURL string) error
	ListTeam(ctx context.Context, organizationID string) ([]entities.User, error)
	InviteMember(ctx context.Context, principal domain.Principal, user entities.User, invitation entities.Invitation) (entities.User, error)
	RevokeMember(ctx context.Context, organizationID, userID string) error

	// --- workspace-aware membership methods ---

	// ListWorkspaces returns every active workspace membership for a user,
	// ordered so that the default/personal workspace appears first.
	ListWorkspaces(ctx context.Context, userID string) ([]entities.WorkspaceSummary, error)

	// Membership returns the membership row for a user within a workspace. It
	// returns ErrNotFound when the user is not an active member.
	Membership(ctx context.Context, userID, organizationID string) (entities.OrganizationMember, error)

	// ByEmailGlobal looks up a user by email across all workspaces (a user is
	// now a global identity), excluding revoked accounts.
	ByEmailGlobal(ctx context.Context, email string) (entities.User, error)

	// CreatePersonalWorkspace provisions a new workspace for an existing user
	// and adds them as the Owner. It also seeds pipeline stages and goals.
	CreatePersonalWorkspace(ctx context.Context, orgName string, user entities.User) (entities.OrganizationMember, error)

	// AddMembership inserts an active membership for a user in a workspace.
	// If an existing (revoked) membership exists it is reactivated instead.
	AddMembership(ctx context.Context, organizationID, userID, role string, invitedBy *string) (entities.OrganizationMember, error)

	// LookupInvitation loads a pending invitation by its token hash.
	LookupInvitation(ctx context.Context, tokenHash string) (entities.Invitation, error)

	// MarkInvitationAccepted records the accepted_at timestamp for an invitation.
	MarkInvitationAccepted(ctx context.Context, invitationID, accepterUserID string) error

	// ActiveMembershipExists reports whether email already has an active
	// membership in the given workspace, returning the conflicting flag.
	ActiveMembershipExists(ctx context.Context, organizationID, email string) (bool, error)
}
