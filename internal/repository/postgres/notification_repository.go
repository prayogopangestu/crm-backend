package postgres

import (
	"context"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/infrastructure/database/postgres"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db       *gorm.DB
	location *time.Location
}

func NewNotificationRepository(db *gorm.DB, location *time.Location) *NotificationRepository {
	return &NotificationRepository{db: db, location: location}
}

func (r *NotificationRepository) List(ctx context.Context, principal domain.Principal) ([]entities.Notification, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT id,title,message,created_at,read_at IS NOT NULL
		FROM notifications
		WHERE organization_id = ? AND (user_id = ? OR user_id IS NULL)
		ORDER BY created_at DESC LIMIT 100`,
		principal.OrganizationID, principal.UserID,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	now := time.Now().In(r.location)
	items := make([]entities.Notification, 0)
	for rows.Next() {
		var item entities.Notification
		if err := rows.Scan(&item.ID, &item.Title, &item.Message, &item.CreatedAt, &item.Read); err != nil {
			return nil, err
		}
		item.Time = postgres.HumanTime(item.CreatedAt, now)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *NotificationRepository) Read(ctx context.Context, principal domain.Principal, id string) error {
	result := r.db.WithContext(ctx).Model(&notificationModel{}).
		Where("id = ? AND organization_id = ? AND (user_id = ? OR user_id IS NULL)", id, principal.OrganizationID, principal.UserID).
		UpdateColumn("read_at", gorm.Expr("COALESCE(read_at, now())"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *NotificationRepository) ReadAll(ctx context.Context, principal domain.Principal) error {
	return r.db.WithContext(ctx).Model(&notificationModel{}).
		Where("organization_id = ? AND (user_id = ? OR user_id IS NULL)", principal.OrganizationID, principal.UserID).
		UpdateColumn("read_at", gorm.Expr("COALESCE(read_at, now())")).Error
}
