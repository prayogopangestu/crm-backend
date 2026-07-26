package postgres

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/infrastructure/database/postgres"
	"gorm.io/gorm"
)

type AnalyticsRepository struct {
	db       *gorm.DB
	location *time.Location
}

func NewAnalyticsRepository(db *gorm.DB, location *time.Location) *AnalyticsRepository {
	return &AnalyticsRepository{db: db, location: location}
}

func (r *AnalyticsRepository) DashboardStats(ctx context.Context, organizationID string, now time.Time) (entities.DashboardStats, error) {
	var stats entities.DashboardStats
	var currentLeads, previousLeads, currentWon, previousWon, revenue int64
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	previousMonth := monthStart.AddDate(0, -1, 0)
	weekStart := analyticsDayStart(now).AddDate(0, 0, -7)
	previousWeek := weekStart.AddDate(0, 0, -7)
	err := r.db.WithContext(ctx).Raw(`
		SELECT
		  count(*) FILTER (WHERE c.created_at >= ?),
		  count(*) FILTER (WHERE c.created_at >= ? AND c.created_at < ?),
		  (SELECT count(*) FROM deals d WHERE d.organization_id=? AND d.deleted_at IS NULL
		    AND d.stage_key='won' AND d.updated_at >= ?),
		  (SELECT count(*) FROM deals d WHERE d.organization_id=? AND d.deleted_at IS NULL
		    AND d.stage_key='won' AND d.updated_at >= ? AND d.updated_at < ?),
		  (SELECT COALESCE(sum(value),0) FROM deals d WHERE d.organization_id=?
		    AND d.deleted_at IS NULL AND d.stage_key='won'),
		  (SELECT count(*) FROM tasks t WHERE t.organization_id=? AND t.deleted_at IS NULL
		    AND t.completed=false AND t.due_date <= ?::date AND t.priority='Tinggi')
		FROM contacts c WHERE c.organization_id=? AND c.deleted_at IS NULL`,
		monthStart, previousMonth, monthStart,
		organizationID, weekStart, organizationID, previousWeek, weekStart,
		organizationID, organizationID, now, organizationID,
	).Row().Scan(
		&currentLeads, &previousLeads, &currentWon, &previousWon, &revenue, &stats.UrgentTasksCount,
	)
	if err != nil {
		return stats, err
	}
	if err := r.db.WithContext(ctx).Raw(
		`SELECT count(*) FROM contacts WHERE organization_id=? AND deleted_at IS NULL`,
		organizationID,
	).Row().Scan(&stats.TotalLeads); err != nil {
		return stats, err
	}
	if err := r.db.WithContext(ctx).Raw(
		`SELECT count(*) FROM deals WHERE organization_id=? AND deleted_at IS NULL AND stage_key='won'`,
		organizationID,
	).Row().Scan(&stats.DealWonCount); err != nil {
		return stats, err
	}
	stats.LeadsTrend = analyticsTrendPercent(currentLeads, previousLeads) + " bulan ini"
	stats.WonTrend = fmt.Sprintf("%+d dari minggu lalu", currentWon-previousWon)
	stats.TotalRevenue = postgres.FormatRupiah(revenue)
	stats.RevenueTrend = "Stabil"
	return stats, nil
}

func (r *AnalyticsRepository) ConversionChart(ctx context.Context, organizationID string, now time.Time) ([]entities.ConversionPoint, error) {
	start := time.Date(now.Year(), now.Month()-5, 1, 0, 0, 0, 0, now.Location())
	rows, err := r.db.WithContext(ctx).Raw(`
		WITH months AS (
		  SELECT generate_series(?::date, date_trunc('month',?::date), interval '1 month') AS month
		)
		SELECT to_char(m.month,'Mon'),
		       CASE WHEN count(d.id)=0 THEN 0
		            ELSE round(100.0*count(d.id) FILTER (WHERE d.stage_key='won')/count(d.id),2)
		       END
		FROM months m
		LEFT JOIN deals d ON d.organization_id=? AND d.deleted_at IS NULL
		  AND date_trunc('month',d.created_at)=m.month
		GROUP BY m.month ORDER BY m.month`,
		start, now, organizationID,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entities.ConversionPoint, 0, 6)
	for rows.Next() {
		var item entities.ConversionPoint
		if err := rows.Scan(&item.Name, &item.Conversion); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AnalyticsRepository) Activities(ctx context.Context, organizationID string, limit int) ([]entities.Activity, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT id,actor_name,action,target,is_highlight,created_at
		FROM activities WHERE organization_id=?
		ORDER BY created_at DESC LIMIT ?`,
		organizationID, limit,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	now := time.Now().In(r.location)
	items := make([]entities.Activity, 0)
	for rows.Next() {
		var item entities.Activity
		if err := rows.Scan(&item.ID, &item.User, &item.Action, &item.Target, &item.IsHighlight, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Time = postgres.HumanTime(item.CreatedAt, now)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AnalyticsRepository) Leaderboard(ctx context.Context, organizationID string, month time.Time) ([]entities.LeaderboardEntry, error) {
	start := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	end := start.AddDate(0, 1, 0)
	previous := start.AddDate(0, -1, 0)
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT trim(u.first_name || ' ' || u.last_name),u.role,u.avatar_url,
		       COALESCE(sum(d.value) FILTER (WHERE d.updated_at >= ? AND d.updated_at < ?),0),
		       COALESCE(sum(d.value) FILTER (WHERE d.updated_at >= ? AND d.updated_at < ?),0)
		FROM users u
		LEFT JOIN deals d ON d.assignee_id=u.id AND d.organization_id=u.organization_id
		  AND d.deleted_at IS NULL AND d.stage_key='won'
		WHERE u.organization_id=? AND u.revoked_at IS NULL
		GROUP BY u.id ORDER BY 4 DESC`,
		start, end, previous, start, organizationID,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entities.LeaderboardEntry, 0)
	for rows.Next() {
		var item entities.LeaderboardEntry
		var previousAmount int64
		if err := rows.Scan(&item.Name, &item.Role, &item.AvatarURL, &item.Amount, &previousAmount); err != nil {
			return nil, err
		}
		item.Rank = len(items) + 1
		trend := analyticsPercentChange(item.Amount, previousAmount)
		item.IsPositive = trend >= 0
		item.Trend = fmt.Sprintf("%+.0f%%", trend)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AnalyticsRepository) LostReasons(ctx context.Context, organizationID string) ([]entities.LostReason, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(NULLIF(lost_reason,''),'Tidak diketahui'),count(*)
		FROM deals
		WHERE organization_id=? AND deleted_at IS NULL AND stage_key='lost'
		GROUP BY 1 ORDER BY 2 DESC`,
		organizationID,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entities.LostReason, 0)
	var total int64
	for rows.Next() {
		var item entities.LostReason
		if err := rows.Scan(&item.Name, &item.Value); err != nil {
			return nil, err
		}
		total += item.Value
		items = append(items, item)
	}
	colors := []string{"#6366f1", "#f59e0b", "#10b981", "#ef4444", "#8b5cf6"}
	for index := range items {
		if total > 0 {
			items[index].Percentage = int64(math.Round(100 * float64(items[index].Value) / float64(total)))
		}
		items[index].Color = colors[index%len(colors)]
	}
	return items, rows.Err()
}

func (r *AnalyticsRepository) Goals(ctx context.Context, organizationID string) ([]entities.PerformanceGoal, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT g.month,g.goal,
		       COALESCE(sum(d.value) FILTER (
		         WHERE d.updated_at >= g.month
		           AND d.updated_at < g.month + interval '1 month'
		       ),0)
		FROM performance_goals g
		LEFT JOIN deals d ON d.organization_id=g.organization_id
		  AND d.deleted_at IS NULL AND d.stage_key='won'
		WHERE g.organization_id=?
		GROUP BY g.id ORDER BY g.month DESC`,
		organizationID,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entities.PerformanceGoal, 0)
	for rows.Next() {
		var month time.Time
		var item entities.PerformanceGoal
		if err := rows.Scan(&month, &item.Goal, &item.Actual); err != nil {
			return nil, err
		}
		item.Month = indonesianMonth(month.Month())
		if item.Goal > 0 {
			item.Percentage = int64(math.Round(100 * float64(item.Actual) / float64(item.Goal)))
		}
		if item.Percentage >= 100 {
			item.Status = fmt.Sprintf("Tercapai (%d%%)", item.Percentage)
		} else {
			item.Status = fmt.Sprintf("Kurang (%d%%)", item.Percentage)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func analyticsDayStart(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func analyticsTrendPercent(current, previous int64) string {
	return fmt.Sprintf("%+.0f%%", analyticsPercentChange(current, previous))
}

func analyticsPercentChange(current, previous int64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return 100 * float64(current-previous) / float64(previous)
}

func indonesianMonth(month time.Month) string {
	names := []string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	return names[int(month)-1]
}
