package postgres

import "time"

type stageModel struct {
	ID             string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string `gorm:"type:uuid"`
	Key            string
	Name           string
	Color          string
	Position       int
	IsSystem       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (stageModel) TableName() string { return "pipeline_stages" }
