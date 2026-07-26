package postgres

import "time"

type goalModel struct {
	ID             string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string    `gorm:"type:uuid"`
	Month          time.Time `gorm:"type:date"`
	Goal           int64
}

func (goalModel) TableName() string { return "performance_goals" }
