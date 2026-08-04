package postgres

import (
	"context"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/infrastructure/database/postgres"
	"gorm.io/gorm"
)

const selectTask = `
	t.id,t.organization_id,t.title,t.company,to_char(t.due_time,'HH24:MI'),
	to_char(t.due_date,'YYYY-MM-DD'),t.type,t.completed,t.notes,t.priority,
	COALESCE(trim(u.first_name || ' ' || u.last_name),''),
	COALESCE(u.id::text,''),t.created_at,t.updated_at`

type TaskRepository struct {
	db       *gorm.DB
	location *time.Location
}

func NewTaskRepository(db *gorm.DB, location *time.Location) *TaskRepository {
	return &TaskRepository{db: db, location: location}
}

func (r *TaskRepository) List(ctx context.Context, organizationID, date, status string, location *time.Location) ([]entities.Task, error) {
	now := time.Now().In(location)
	query := r.db.WithContext(ctx).Table("tasks AS t").Select(selectTask).
		Joins("LEFT JOIN organization_members om ON om.user_id = t.assignee_id AND om.organization_id = t.organization_id AND om.revoked_at IS NULL").
		Joins("LEFT JOIN users u ON u.id = t.assignee_id").
		Where("t.organization_id = ? AND t.deleted_at IS NULL", organizationID)
	if date != "" {
		query = query.Where("t.due_date = ?::date", date)
	} else {
		today := now.Format("2006-01-02")
		switch status {
		case "overdue":
			query = query.Where("t.due_date < ?::date AND t.completed = false", today)
		case "today":
			query = query.Where("t.due_date = ?::date", today)
		case "upcoming":
			query = query.Where("t.due_date > ?::date", today)
		}
	}
	rows, err := query.Order("t.due_date,t.due_time").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entities.Task, 0)
	for rows.Next() {
		item, err := scanTask(rows, now)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *TaskRepository) Create(ctx context.Context, principal domain.Principal, input entities.Task) (entities.Task, error) {
	assigneeID, err := r.resolveAssignee(ctx, principal, input.AssigneeID, input.Assignee)
	if err != nil {
		return entities.Task{}, err
	}
	dueDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return entities.Task{}, domain.ErrInvalidInput
	}
	record := taskModel{
		OrganizationID: principal.OrganizationID, AssigneeID: &assigneeID,
		Title: input.Title, Company: input.Company, DueDate: dueDate, DueTime: input.Time,
		Type: input.Type, Priority: input.Priority, Notes: input.Notes, Completed: input.Completed,
	}
	if input.Completed {
		now := time.Now()
		record.CompletedAt = &now
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return entities.Task{}, postgres.MapError(err)
	}
	return r.byID(ctx, principal.OrganizationID, record.ID)
}

func (r *TaskRepository) Update(ctx context.Context, principal domain.Principal, id string, input entities.Task) (entities.Task, error) {
	updates := map[string]any{"updated_at": time.Now()}
	if input.Title != "" {
		updates["title"] = input.Title
	}
	if input.Company != "" {
		updates["company"] = input.Company
	}
	if input.Date != "" {
		dueDate, err := time.Parse("2006-01-02", input.Date)
		if err != nil {
			return entities.Task{}, domain.ErrInvalidInput
		}
		updates["due_date"] = dueDate
	}
	if input.Time != "" {
		updates["due_time"] = input.Time
	}
	if input.Type != "" {
		updates["type"] = input.Type
	}
	if input.Priority != "" {
		updates["priority"] = input.Priority
	}
	if input.Notes != "" {
		updates["notes"] = input.Notes
	}
	if input.AssigneeID != "" || input.Assignee != "" {
		assigneeID, err := r.resolveAssignee(ctx, principal, input.AssigneeID, input.Assignee)
		if err != nil {
			return entities.Task{}, err
		}
		updates["assignee_id"] = assigneeID
	}
	result := r.db.WithContext(ctx).Model(&taskModel{}).
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, principal.OrganizationID).
		Updates(updates)
	if result.Error != nil {
		return entities.Task{}, postgres.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return entities.Task{}, domain.ErrNotFound
	}
	return r.byID(ctx, principal.OrganizationID, id)
}

func (r *TaskRepository) Toggle(ctx context.Context, principal domain.Principal, id string, completed bool) error {
	var completedAt any
	if completed {
		completedAt = time.Now()
	}
	result := r.db.WithContext(ctx).Model(&taskModel{}).
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, principal.OrganizationID).
		Updates(map[string]any{"completed": completed, "completed_at": completedAt, "updated_at": time.Now()})
	if result.Error != nil {
		return postgres.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, principal domain.Principal, id string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&taskModel{}).
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

func (r *TaskRepository) byID(ctx context.Context, organizationID, id string) (entities.Task, error) {
	item, err := scanTask(r.db.WithContext(ctx).Table("tasks AS t").Select(selectTask).
		Joins("LEFT JOIN organization_members om ON om.user_id = t.assignee_id AND om.organization_id = t.organization_id AND om.revoked_at IS NULL").
		Joins("LEFT JOIN users u ON u.id = t.assignee_id").
		Where("t.id = ? AND t.organization_id = ? AND t.deleted_at IS NULL", id, organizationID).Row(),
		time.Now().In(r.location))
	return item, postgres.MapError(err)
}

func scanTask(row postgres.Scanner, now time.Time) (entities.Task, error) {
	var item entities.Task
	err := row.Scan(
		&item.ID, &item.OrganizationID, &item.Title, &item.Company, &item.Time,
		&item.Date, &item.Type, &item.Completed, &item.Notes, &item.Priority,
		&item.Assignee, &item.AssigneeID, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return entities.Task{}, err
	}
	due, _ := time.ParseInLocation("2006-01-02", item.Date, now.Location())
	switch {
	case due.Before(dayStart(now)) && !item.Completed:
		item.Status = "overdue"
	case due.After(dayStart(now)):
		item.Status = "upcoming"
	default:
		item.Status = "today"
	}
	return item, nil
}

func (r *TaskRepository) resolveAssignee(ctx context.Context, principal domain.Principal, id, name string) (string, error) {
	if id == "" && name == "" {
		return principal.UserID, nil
	}
	var resolved string
	query := r.db.WithContext(ctx).Table("organization_members AS om").
		Select("om.user_id").
		Joins("JOIN users u ON u.id = om.user_id").
		Where("om.organization_id = ? AND om.revoked_at IS NULL AND om.status = 'Aktif' AND om.role IN (?, ?, ?)",
			principal.OrganizationID, domain.RoleOwner, domain.RoleAdmin, domain.RoleSales)
	if id != "" {
		query = query.Where("om.user_id = ?", id)
	} else {
		query = query.Where("lower(trim(u.first_name || ' ' || u.last_name)) = lower(?)", name)
	}
	if err := query.Row().Scan(&resolved); err != nil {
		return "", postgres.MapError(err)
	}
	return resolved, nil
}

func dayStart(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}
