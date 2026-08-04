package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/infrastructure/database/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const selectDeal = `
	SELECT d.id,d.organization_id,d.title,d.company,d.value,d.priority,d.stage_key,d.lost_reason,
	       COALESCE(u.id::text,''),COALESCE(trim(u.first_name || ' ' || u.last_name),''),
	       COALESCE(u.avatar_url,''),d.created_at,d.updated_at
	FROM deals d
	LEFT JOIN organization_members om ON om.user_id = d.assignee_id AND om.organization_id = d.organization_id AND om.revoked_at IS NULL
	LEFT JOIN users u ON u.id = d.assignee_id`

type DealRepository struct {
	db       *gorm.DB
	location *time.Location
}

func NewDealRepository(db *gorm.DB, location *time.Location) *DealRepository {
	return &DealRepository{db: db, location: location}
}

func (r *DealRepository) List(ctx context.Context, organizationID string) ([]entities.Deal, error) {
	rows, err := r.db.WithContext(ctx).Raw(
		selectDeal+` WHERE d.organization_id = ? AND d.deleted_at IS NULL ORDER BY d.created_at`,
		organizationID,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entities.Deal, 0)
	for rows.Next() {
		item, err := scanDeal(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *DealRepository) Create(ctx context.Context, principal domain.Principal, input entities.Deal) (entities.Deal, error) {
	if err := r.ensureStage(ctx, principal.OrganizationID, input.Stage); err != nil {
		return entities.Deal{}, err
	}
	assigneeID := input.Assignee.ID
	if assigneeID == "" {
		assigneeID = principal.UserID
	}
	if err := r.ensureUser(ctx, principal.OrganizationID, assigneeID); err != nil {
		return entities.Deal{}, err
	}
	record := dealModel{
		OrganizationID: principal.OrganizationID, AssigneeID: &assigneeID,
		Title: input.Title, Company: input.Company, Value: input.Value,
		Priority: input.Priority, StageKey: input.Stage, LostReason: input.LostReason,
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return entities.Deal{}, postgres.MapError(err)
	}
	return r.byID(ctx, principal.OrganizationID, record.ID)
}

func (r *DealRepository) Update(ctx context.Context, principal domain.Principal, id string, input entities.Deal) (entities.Deal, error) {
	if err := r.ensureStage(ctx, principal.OrganizationID, input.Stage); err != nil {
		return entities.Deal{}, err
	}
	assigneeID := input.Assignee.ID
	if assigneeID == "" {
		assigneeID = principal.UserID
	}
	if err := r.ensureUser(ctx, principal.OrganizationID, assigneeID); err != nil {
		return entities.Deal{}, err
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current dealModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, principal.OrganizationID).
			First(&current).Error; err != nil {
			return postgres.MapError(err)
		}
		if err := tx.Model(&current).Updates(map[string]any{
			"title": input.Title, "company": input.Company, "value": input.Value,
			"priority": input.Priority, "stage_key": input.Stage,
			"lost_reason": input.LostReason, "assignee_id": assigneeID, "updated_at": time.Now(),
		}).Error; err != nil {
			return postgres.MapError(err)
		}
		if current.StageKey != input.Stage {
			return r.recordStageChange(tx, principal, input.Title, input.Stage)
		}
		return nil
	})
	if err != nil {
		return entities.Deal{}, err
	}
	return r.byID(ctx, principal.OrganizationID, id)
}

func (r *DealRepository) UpdateStage(ctx context.Context, principal domain.Principal, id, stage, lostReason string) error {
	if err := r.ensureStage(ctx, principal.OrganizationID, stage); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current dealModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, principal.OrganizationID).
			First(&current).Error; err != nil {
			return postgres.MapError(err)
		}
		if err := tx.Model(&current).Updates(map[string]any{
			"stage_key": stage, "lost_reason": lostReason, "updated_at": time.Now(),
		}).Error; err != nil {
			return postgres.MapError(err)
		}
		if current.StageKey == stage {
			return nil
		}
		return r.recordStageChange(tx, principal, current.Title, stage)
	})
}

func (r *DealRepository) Delete(ctx context.Context, principal domain.Principal, id string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&dealModel{}).
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, principal.OrganizationID).
		Updates(map[string]any{"deleted_at": now, "updated_at": now})
	if result.Error != nil {
		return postgres.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *DealRepository) recordStageChange(tx *gorm.DB, principal domain.Principal, title, stage string) error {
	if err := tx.Create(&activityModel{
		OrganizationID: principal.OrganizationID, ActorID: principal.UserID,
		ActorName: principal.Name, Action: "memindahkan deal ke " + stage,
		Target: title, IsHighlight: stage == "won",
	}).Error; err != nil {
		return postgres.MapError(err)
	}
	if stage != "won" {
		return nil
	}
	message := "Deal " + title + " berhasil dimenangkan oleh " + principal.Name
	if err := tx.Exec(`
		INSERT INTO notifications (organization_id,user_id,title,message)
		SELECT ?,om.user_id,'Deal Won!',? FROM organization_members om
		WHERE om.organization_id = ? AND om.revoked_at IS NULL AND om.status = 'Aktif'`,
		principal.OrganizationID, message, principal.OrganizationID,
	).Error; err != nil {
		return postgres.MapError(err)
	}
	payload, err := json.Marshal(map[string]string{"message": message})
	if err != nil {
		return err
	}
	return postgres.MapError(tx.Create(&outboxModel{
		OrganizationID: principal.OrganizationID, EventType: "telegram.deal_won", Payload: payload,
	}).Error)
}

func (r *DealRepository) byID(ctx context.Context, organizationID, id string) (entities.Deal, error) {
	item, err := scanDeal(r.db.WithContext(ctx).Raw(
		selectDeal+` WHERE d.id = ? AND d.organization_id = ? AND d.deleted_at IS NULL`,
		id, organizationID,
	).Row())
	return item, postgres.MapError(err)
}

func scanDeal(row postgres.Scanner) (entities.Deal, error) {
	var item entities.Deal
	err := row.Scan(
		&item.ID, &item.OrganizationID, &item.Title, &item.Company, &item.Value,
		&item.Priority, &item.Stage, &item.LostReason, &item.Assignee.ID,
		&item.Assignee.Name, &item.Assignee.AvatarURL, &item.CreatedAt, &item.UpdatedAt,
	)
	return item, err
}

func (r *DealRepository) ensureStage(ctx context.Context, organizationID, key string) error {
	var count int64
	if err := r.db.WithContext(ctx).Table("pipeline_stages").
		Where("organization_id = ? AND key = ?", organizationID, key).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return domain.ErrInvalidInput
	}
	return nil
}

func (r *DealRepository) ensureUser(ctx context.Context, organizationID, userID string) error {
	var count int64
	if err := r.db.WithContext(ctx).Table("organization_members").
		Where("organization_id = ? AND user_id = ? AND status = 'Aktif' AND revoked_at IS NULL AND role IN (?, ?, ?)",
			organizationID, userID, domain.RoleOwner, domain.RoleAdmin, domain.RoleSales).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return domain.ErrInvalidInput
	}
	return nil
}
