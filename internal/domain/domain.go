package domain

import (
	"context"
	"time"

	domainerrors "github.com/prayogopangestu/crm-system/backend/internal/domain/errors"
)

// Sentinel errors are defined in the dedicated sub-package
// internal/domain/errors and re-exported here so existing callers can keep
// using `domain.ErrXxx`. New code should import the errors package directly.
var (
	ErrNotFound       = domainerrors.ErrNotFound
	ErrConflict       = domainerrors.ErrConflict
	ErrUnauthorized   = domainerrors.ErrUnauthorized
	ErrForbidden      = domainerrors.ErrForbidden
	ErrInvalidInput   = domainerrors.ErrInvalidInput
	ErrStageInUse     = domainerrors.ErrStageInUse
	ErrInviteExpired  = domainerrors.ErrInviteExpired
	ErrInviteUsed     = domainerrors.ErrInviteUsed
	ErrRedisRateLimit = domainerrors.ErrRedisRateLimit
)

const (
	RoleOwner  = "Owner"
	RoleAdmin  = "Admin"
	RoleSales  = "Staf Sales"
	RoleViewer = "Viewer"
)

// adminRoles are the roles that can manage team members, integrations and
// pipeline configuration within a workspace.
var adminRoles = map[string]bool{RoleOwner: true, RoleAdmin: true}

// writeCRMRoles are the roles that may create, update or delete CRM data.
var writeCRMRoles = map[string]bool{RoleOwner: true, RoleAdmin: true, RoleSales: true}

// readCRMRoles are the roles that may read CRM data. All membership roles can
// read by default.
var readCRMRoles = map[string]bool{RoleOwner: true, RoleAdmin: true, RoleSales: true, RoleViewer: true}

type Principal struct {
	UserID         string
	OrganizationID string
	Role           string
	Name           string
}

type principalKey struct{}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalKey{}).(Principal)
	return principal, ok
}

// RequireAdmin keeps backwards compatibility: the historical "team admin"
// gate is now expressed as "workspace admin" (Owner or Admin).
func RequireAdmin(principal Principal) error {
	if !adminRoles[principal.Role] {
		return ErrForbidden
	}
	return nil
}

// RequireWorkspaceAdmin allows only workspace owners and admins.
func RequireWorkspaceAdmin(principal Principal) error {
	if !adminRoles[principal.Role] {
		return ErrForbidden
	}
	return nil
}

// RequireOwner allows only the workspace owner (transfer/delete workspace).
func RequireOwner(principal Principal) error {
	if principal.Role != RoleOwner {
		return ErrForbidden
	}
	return nil
}

// RequireCanWriteCRM allows roles that may mutate CRM data.
func RequireCanWriteCRM(principal Principal) error {
	if !writeCRMRoles[principal.Role] {
		return ErrForbidden
	}
	return nil
}

// RequireCanReadCRM allows any active membership role to read CRM data.
func RequireCanReadCRM(principal Principal) error {
	if !readCRMRoles[principal.Role] {
		return ErrForbidden
	}
	return nil
}

// ValidMembershipRole reports whether role is one of the membership roles.
func ValidMembershipRole(role string) bool {
	return role == RoleOwner || role == RoleAdmin || role == RoleSales || role == RoleViewer
}

// InvitableRole reports whether role may be assigned through an invitation.
// Owners are created automatically (on register/personal workspace) and must
// not be granted via invite.
func InvitableRole(role string) bool {
	return role == RoleAdmin || role == RoleSales || role == RoleViewer
}

type Cache interface {
	GetJSON(ctx context.Context, key string, dst any) (bool, error)
	SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error
	DeletePattern(ctx context.Context, pattern string) error
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
	Close() error
}

type TokenManager interface {
	Create(userID, organizationID, role, name string) (string, error)
	Parse(raw string) (Principal, error)
}
