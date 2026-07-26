package postgres

import "time"

type taskModel struct {
	ID             string  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string  `gorm:"type:uuid"`
	AssigneeID     *string `gorm:"type:uuid"`
	Title          string
	Company        string
	DueDate        time.Time `gorm:"type:date"`
	DueTime        string    `gorm:"type:time"`
	Type           string
	Priority       string
	Notes          string
	Completed      bool
	CompletedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

func (taskModel) TableName() string { return "tasks" }
