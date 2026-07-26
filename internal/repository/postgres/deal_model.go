package postgres

import "time"

type dealModel struct {
	ID             string  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string  `gorm:"type:uuid"`
	AssigneeID     *string `gorm:"type:uuid"`
	Title          string
	Company        string
	Value          int64
	Priority       string
	StageKey       string
	LostReason     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

func (dealModel) TableName() string { return "deals" }
