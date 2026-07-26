package repositories

import (
	"context"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
)

type UserRepository interface {
	Register(ctx context.Context, orgName string, user entities.User) (entities.User, error)
	ByEmail(ctx context.Context, email string) (entities.User, error)
	ByID(ctx context.Context, organizationID, userID string) (entities.User, error)
	UpdateProfile(ctx context.Context, principal domain.Principal, firstName, lastName, email string) (entities.User, error)
	AcceptInvite(ctx context.Context, tokenHash, passwordHash string) (entities.User, error)
	ListTeam(ctx context.Context, organizationID string) ([]entities.User, error)
	InviteMember(ctx context.Context, principal domain.Principal, user entities.User, invitation entities.Invitation) (entities.User, error)
	RevokeMember(ctx context.Context, organizationID, userID string) error
}
