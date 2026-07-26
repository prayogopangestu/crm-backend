package search

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/repositories"
	"github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
)

type Service struct {
	repository repositories.SearchRepository
	cache      support.CacheHelper
}

func NewService(repository repositories.SearchRepository, cache support.CacheHelper) *Service {
	return &Service{repository: repository, cache: cache}
}

func (s *Service) Search(ctx context.Context, principal domain.Principal, query string) (entities.SearchResult, error) {
	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return entities.SearchResult{}, domain.ErrInvalidInput
	}
	sum := sha256.Sum256([]byte(strings.ToLower(query)))
	key := "crm:" + principal.OrganizationID + ":search:" + hex.EncodeToString(sum[:8])
	var result entities.SearchResult
	err := s.cache.Load(ctx, key, 30*time.Second, &result, func() error {
		var err error
		result, err = s.repository.Search(ctx, principal.OrganizationID, query)
		return err
	})
	return result, err
}
