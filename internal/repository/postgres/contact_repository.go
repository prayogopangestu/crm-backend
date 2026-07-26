package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/infrastructure/database/postgres"
	"gorm.io/gorm"
)

type ContactRepository struct {
	db       *gorm.DB
	location *time.Location
}

func NewContactRepository(db *gorm.DB, location *time.Location) *ContactRepository {
	return &ContactRepository{db: db, location: location}
}

func (r *ContactRepository) List(ctx context.Context, organizationID, search, status string, page entities.ContactPage) (entities.ContactList, error) {
	query := r.db.WithContext(ctx).Model(&contactModel{}).Where("organization_id = ? AND deleted_at IS NULL", organizationID)
	if search != "" {
		pattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ? OR company ILIKE ?", pattern, pattern, pattern)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return entities.ContactList{}, err
	}
	var records []contactModel
	if err := query.Order("created_at DESC").Limit(page.Limit).Offset((page.Page - 1) * page.Limit).Find(&records).Error; err != nil {
		return entities.ContactList{}, err
	}
	items := make([]entities.Contact, 0, len(records))
	for _, record := range records {
		items = append(items, r.toEntity(record))
	}
	return entities.ContactList{Data: items, Total: total, Page: page.Page}, nil
}

func (r *ContactRepository) Create(ctx context.Context, principal domain.Principal, input entities.Contact) (entities.Contact, error) {
	var item entities.Contact
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ownerID := principal.UserID
		record := contactModel{
			OrganizationID: principal.OrganizationID, OwnerID: &ownerID,
			Name: input.Name, Email: strings.ToLower(input.Email), Company: input.Company,
			Role: input.Role, Status: input.Status, AvatarURL: input.AvatarURL, LastContactedAt: time.Now(),
		}
		if err := tx.Create(&record).Error; err != nil {
			return postgres.MapError(err)
		}
		if err := tx.Create(&activityModel{
			OrganizationID: principal.OrganizationID, ActorID: principal.UserID,
			ActorName: principal.Name, Action: "menambahkan kontak baru", Target: record.Name,
		}).Error; err != nil {
			return postgres.MapError(err)
		}
		item = r.toEntity(record)
		return nil
	})
	return item, err
}

func (r *ContactRepository) Update(ctx context.Context, principal domain.Principal, id string, input entities.Contact) (entities.Contact, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&contactModel{}).
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, principal.OrganizationID).
		Updates(map[string]any{
			"name": input.Name, "email": strings.ToLower(input.Email), "company": input.Company,
			"role": input.Role, "status": input.Status, "avatar_url": input.AvatarURL,
			"last_contacted_at": now, "updated_at": now,
		})
	if result.Error != nil {
		return entities.Contact{}, postgres.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return entities.Contact{}, domain.ErrNotFound
	}
	var record contactModel
	if err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, principal.OrganizationID).
		First(&record).Error; err != nil {
		return entities.Contact{}, postgres.MapError(err)
	}
	return r.toEntity(record), nil
}

func (r *ContactRepository) Delete(ctx context.Context, principal domain.Principal, id string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&contactModel{}).
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

func (r *ContactRepository) toEntity(record contactModel) entities.Contact {
	ownerID := ""
	if record.OwnerID != nil {
		ownerID = *record.OwnerID
	}
	item := entities.Contact{
		ID: record.ID, OrganizationID: record.OrganizationID, OwnerID: ownerID,
		Name: record.Name, Email: record.Email, Company: record.Company, Role: record.Role,
		Status: record.Status, AvatarURL: record.AvatarURL, LastContactedAt: record.LastContactedAt,
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}
	item.Initials = postgres.Initials(item.Name)
	item.LastContacted = postgres.HumanTime(item.LastContactedAt, time.Now().In(r.location))
	return item
}
