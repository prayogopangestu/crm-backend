package repositories

import (
	"context"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
)

type NotificationRepository interface {
	List(ctx context.Context, principal domain.Principal) ([]entities.Notification, error)
	Read(ctx context.Context, principal domain.Principal, id string) error
	ReadAll(ctx context.Context, principal domain.Principal) error
}
