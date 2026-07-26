package repositories

import (
	"context"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
)

type TaskRepository interface {
	List(ctx context.Context, organizationID, date, status string, location *time.Location) ([]entities.Task, error)
	Create(ctx context.Context, principal domain.Principal, task entities.Task) (entities.Task, error)
	Update(ctx context.Context, principal domain.Principal, id string, task entities.Task) (entities.Task, error)
	Toggle(ctx context.Context, principal domain.Principal, id string, completed bool) error
	Delete(ctx context.Context, principal domain.Principal, id string) error
}
