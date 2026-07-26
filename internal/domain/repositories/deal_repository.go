package repositories

import (
	"context"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
)

type DealRepository interface {
	List(ctx context.Context, organizationID string) ([]entities.Deal, error)
	Create(ctx context.Context, principal domain.Principal, deal entities.Deal) (entities.Deal, error)
	Update(ctx context.Context, principal domain.Principal, id string, deal entities.Deal) (entities.Deal, error)
	UpdateStage(ctx context.Context, principal domain.Principal, id, stage, lostReason string) error
	Delete(ctx context.Context, principal domain.Principal, id string) error
}
