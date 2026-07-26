package repositories

import (
	"context"

	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
)

type SearchRepository interface {
	Search(ctx context.Context, organizationID, query string) (entities.SearchResult, error)
}
