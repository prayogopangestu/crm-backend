package repositories

import (
	"context"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
)

type ContactRepository interface {
	List(ctx context.Context, organizationID, search, status string, page entities.ContactPage) (entities.ContactList, error)
	Create(ctx context.Context, principal domain.Principal, contact entities.Contact) (entities.Contact, error)
	Update(ctx context.Context, principal domain.Principal, id string, contact entities.Contact) (entities.Contact, error)
	Delete(ctx context.Context, principal domain.Principal, id string) error
}
