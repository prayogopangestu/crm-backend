package repositories

import (
	"context"

	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
)

type PipelineRepository interface {
	List(ctx context.Context, organizationID string) ([]entities.Stage, error)
	Create(ctx context.Context, organizationID string, stage entities.Stage) (entities.Stage, error)
	Reorder(ctx context.Context, organizationID string, ids []string) error
	Delete(ctx context.Context, organizationID, id string) error
}
