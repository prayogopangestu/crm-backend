package support

import (
	"context"
	"fmt"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
)

type Logger interface {
	Warn(msg string, args ...any)
}

type CacheHelper struct {
	Cache  domain.Cache
	Logger Logger
}

func (h CacheHelper) Load(ctx context.Context, key string, ttl time.Duration, dst any, loader func() error) error {
	if h.Cache != nil {
		hit, err := h.Cache.GetJSON(ctx, key, dst)
		if err != nil {
			h.Logger.Warn("cache read failed", "key", key, "error", err)
		} else if hit {
			return nil
		}
	}
	if err := loader(); err != nil {
		return err
	}
	if h.Cache != nil {
		if err := h.Cache.SetJSON(ctx, key, dst, ttl); err != nil {
			h.Logger.Warn("cache write failed", "key", key, "error", err)
		}
	}
	return nil
}

func (h CacheHelper) InvalidateCRM(ctx context.Context, organizationID string) {
	if h.Cache == nil {
		return
	}
	for _, pattern := range []string{
		fmt.Sprintf("crm:%s:dashboard:*", organizationID),
		fmt.Sprintf("crm:%s:reports:*", organizationID),
		fmt.Sprintf("crm:%s:search:*", organizationID),
	} {
		if err := h.Cache.DeletePattern(ctx, pattern); err != nil {
			h.Logger.Warn("cache invalidation failed", "pattern", pattern, "error", err)
		}
	}
}

func (h CacheHelper) InvalidateProfile(ctx context.Context, organizationID, userID string) {
	if h.Cache == nil {
		return
	}
	key := fmt.Sprintf("crm:%s:profile:%s", organizationID, userID)
	if err := h.Cache.DeletePattern(ctx, key); err != nil {
		h.Logger.Warn("profile cache invalidation failed", "key", key, "error", err)
	}
}
