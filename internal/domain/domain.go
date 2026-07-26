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
	RoleAdmin = "Admin"
	RoleSales = "Staf Sales"
)

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

func RequireAdmin(principal Principal) error {
	if principal.Role != RoleAdmin {
		return ErrForbidden
	}
	return nil
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
