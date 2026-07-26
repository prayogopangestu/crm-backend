package repositories

import (
	"context"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
)

type AnalyticsRepository interface {
	DashboardStats(ctx context.Context, organizationID string, now time.Time) (entities.DashboardStats, error)
	ConversionChart(ctx context.Context, organizationID string, now time.Time) ([]entities.ConversionPoint, error)
	Activities(ctx context.Context, organizationID string, limit int) ([]entities.Activity, error)
	Leaderboard(ctx context.Context, organizationID string, month time.Time) ([]entities.LeaderboardEntry, error)
	LostReasons(ctx context.Context, organizationID string) ([]entities.LostReason, error)
	Goals(ctx context.Context, organizationID string) ([]entities.PerformanceGoal, error)
}
