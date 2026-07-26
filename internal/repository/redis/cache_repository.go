package redis

import (
	"context"
	"time"
)

type CacheRepository struct{}

func NewCacheRepository() *CacheRepository {
	return &CacheRepository{}
}

func (r *CacheRepository) GetJSON(ctx context.Context, key string, dst any) (bool, error) {
	return false, nil
}

func (r *CacheRepository) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	return nil
}

func (r *CacheRepository) DeletePattern(ctx context.Context, pattern string) error {
	return nil
}

func (r *CacheRepository) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	return true, nil
}

func (r *CacheRepository) Close() error {
	return nil
}
