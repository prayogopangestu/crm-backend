package postgres

import (
	"context"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/infrastructure/database/postgres"
	"gorm.io/gorm"
)

type SearchRepository struct {
	db       *gorm.DB
	location *time.Location
}

func NewSearchRepository(db *gorm.DB, location *time.Location) *SearchRepository {
	return &SearchRepository{db: db, location: location}
}

func (r *SearchRepository) Search(ctx context.Context, organizationID, query string) (entities.SearchResult, error) {
	pattern := "%" + query + "%"
	result := entities.SearchResult{Contacts: []entities.Contact{}, Tasks: []entities.Task{}, Deals: []entities.Deal{}}

	contactRows, err := r.db.WithContext(ctx).Raw(`
		SELECT id,organization_id,COALESCE(owner_id::text,''),name,email,company,role,status,avatar_url,
		       last_contacted_at,created_at,updated_at
		FROM contacts
		WHERE organization_id = ? AND deleted_at IS NULL
		  AND (name ILIKE ? OR email ILIKE ? OR company ILIKE ?)
		ORDER BY updated_at DESC LIMIT 10`,
		organizationID, pattern, pattern, pattern,
	).Rows()
	if err != nil {
		return result, err
	}
	for contactRows.Next() {
		var item entities.Contact
		if err := contactRows.Scan(
			&item.ID, &item.OrganizationID, &item.OwnerID, &item.Name, &item.Email,
			&item.Company, &item.Role, &item.Status, &item.AvatarURL,
			&item.LastContactedAt, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			contactRows.Close()
			return result, err
		}
		item.Initials = postgres.Initials(item.Name)
		item.LastContacted = postgres.HumanTime(item.LastContactedAt, time.Now().In(r.location))
		result.Contacts = append(result.Contacts, item)
	}
	contactRows.Close()

	taskRows, err := r.db.WithContext(ctx).Raw(`
		SELECT t.id,t.organization_id,t.title,t.company,to_char(t.due_time,'HH24:MI'),
		       to_char(t.due_date,'YYYY-MM-DD'),t.type,t.completed,t.notes,t.priority,
		       COALESCE(trim(u.first_name || ' ' || u.last_name),''),
		       COALESCE(u.id::text,''),t.created_at,t.updated_at
		FROM tasks t
		LEFT JOIN organization_members om ON om.user_id=t.assignee_id AND om.organization_id=t.organization_id AND om.status='Aktif' AND om.revoked_at IS NULL
		LEFT JOIN users u ON u.id=om.user_id AND u.revoked_at IS NULL
		WHERE t.organization_id = ? AND t.deleted_at IS NULL
		  AND (t.title ILIKE ? OR t.company ILIKE ? OR t.notes ILIKE ?)
		ORDER BY t.updated_at DESC LIMIT 10`,
		organizationID, pattern, pattern, pattern,
	).Rows()
	if err != nil {
		return result, err
	}
	now := time.Now().In(r.location)
	for taskRows.Next() {
		var item entities.Task
		if err := taskRows.Scan(
			&item.ID, &item.OrganizationID, &item.Title, &item.Company, &item.Time,
			&item.Date, &item.Type, &item.Completed, &item.Notes, &item.Priority,
			&item.Assignee, &item.AssigneeID, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			taskRows.Close()
			return result, err
		}
		due, _ := time.ParseInLocation("2006-01-02", item.Date, now.Location())
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		switch {
		case due.Before(today) && !item.Completed:
			item.Status = "overdue"
		case due.After(today):
			item.Status = "upcoming"
		default:
			item.Status = "today"
		}
		result.Tasks = append(result.Tasks, item)
	}
	taskRows.Close()

	dealRows, err := r.db.WithContext(ctx).Raw(`
		SELECT d.id,d.organization_id,d.title,d.company,d.value,d.priority,d.stage_key,d.lost_reason,
		       COALESCE(u.id::text,''),COALESCE(trim(u.first_name || ' ' || u.last_name),''),
		       COALESCE(u.avatar_url,''),d.created_at,d.updated_at
		FROM deals d
		LEFT JOIN organization_members om ON om.user_id=d.assignee_id AND om.organization_id=d.organization_id AND om.status='Aktif' AND om.revoked_at IS NULL
		LEFT JOIN users u ON u.id=om.user_id AND u.revoked_at IS NULL
		WHERE d.organization_id = ? AND d.deleted_at IS NULL
		  AND (d.title ILIKE ? OR d.company ILIKE ?)
		ORDER BY d.updated_at DESC LIMIT 10`,
		organizationID, pattern, pattern,
	).Rows()
	if err != nil {
		return result, err
	}
	defer dealRows.Close()
	for dealRows.Next() {
		var item entities.Deal
		if err := dealRows.Scan(
			&item.ID, &item.OrganizationID, &item.Title, &item.Company, &item.Value,
			&item.Priority, &item.Stage, &item.LostReason, &item.Assignee.ID,
			&item.Assignee.Name, &item.Assignee.AvatarURL, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return result, err
		}
		result.Deals = append(result.Deals, item)
	}
	return result, dealRows.Err()
}
