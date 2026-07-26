package postgres

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/infrastructure/database/postgres"
	"gorm.io/gorm"
)

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

type PipelineRepository struct {
	db       *gorm.DB
	location *time.Location
}

func NewPipelineRepository(db *gorm.DB, location *time.Location) *PipelineRepository {
	return &PipelineRepository{db: db, location: location}
}

func (r *PipelineRepository) List(ctx context.Context, organizationID string) ([]entities.Stage, error) {
	var records []stageModel
	if err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Order("position").Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]entities.Stage, 0, len(records))
	for _, record := range records {
		items = append(items, toStageEntity(record))
	}
	return items, nil
}

func (r *PipelineRepository) Create(ctx context.Context, organizationID string, stage entities.Stage) (entities.Stage, error) {
	stage.Key = strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(stage.Name), "-"), "-")
	if stage.Key == "" {
		return entities.Stage{}, domain.ErrInvalidInput
	}
	if stage.Color == "" {
		stage.Color = "bg-surface-variant"
	}
	var record stageModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var maxPosition int
		if err := tx.Model(&stageModel{}).Select("COALESCE(MAX(position), 0)").
			Where("organization_id = ?", organizationID).Scan(&maxPosition).Error; err != nil {
			return err
		}
		record = stageModel{
			OrganizationID: organizationID, Key: stage.Key, Name: stage.Name,
			Color: stage.Color, Position: maxPosition + 1,
		}
		return postgres.MapError(tx.Create(&record).Error)
	})
	return toStageEntity(record), err
}

func (r *PipelineRepository) Reorder(ctx context.Context, organizationID string, ids []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&stageModel{}).Where("organization_id = ?", organizationID).Count(&count).Error; err != nil {
			return err
		}
		if count != int64(len(ids)) {
			return domain.ErrInvalidInput
		}
		if err := tx.Model(&stageModel{}).Where("organization_id = ?", organizationID).
			UpdateColumn("position", gorm.Expr("position + 1000")).Error; err != nil {
			return postgres.MapError(err)
		}
		for index, id := range ids {
			result := tx.Model(&stageModel{}).Where("id = ? AND organization_id = ?", id, organizationID).
				Updates(map[string]any{"position": index + 1, "updated_at": time.Now()})
			if result.Error != nil {
				return postgres.MapError(result.Error)
			}
			if result.RowsAffected != 1 {
				return domain.ErrInvalidInput
			}
		}
		return nil
	})
}

func (r *PipelineRepository) Delete(ctx context.Context, organizationID, id string) error {
	var record stageModel
	if err := r.db.WithContext(ctx).Where("id = ? AND organization_id = ?", id, organizationID).First(&record).Error; err != nil {
		return postgres.MapError(err)
	}
	if record.IsSystem {
		return domain.ErrForbidden
	}
	var used int64
	if err := r.db.WithContext(ctx).Table("deals").
		Where("organization_id = ? AND stage_key = ? AND deleted_at IS NULL", organizationID, record.Key).
		Count(&used).Error; err != nil {
		return err
	}
	if used > 0 {
		return domain.ErrStageInUse
	}
	result := r.db.WithContext(ctx).Where("id = ? AND organization_id = ?", id, organizationID).Delete(&stageModel{})
	if result.Error != nil {
		return postgres.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func toStageEntity(record stageModel) entities.Stage {
	return entities.Stage{
		ID: record.ID, OrganizationID: record.OrganizationID, Key: record.Key,
		Name: record.Name, Color: record.Color, Position: record.Position,
		IsSystem: record.IsSystem, CreatedAt: record.CreatedAt,
	}
}
