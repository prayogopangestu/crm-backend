package entities

import "time"

type Stage struct {
	ID             string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string    `json:"-" gorm:"type:uuid;not null;column:organization_id"`
	Key            string    `json:"key" gorm:"type:text;not null"`
	Name           string    `json:"name" gorm:"type:text;not null"`
	Color          string    `json:"color" gorm:"type:text;not null;default:'bg-surface-variant'"`
	Position       int       `json:"position" gorm:"type:integer;not null"`
	IsSystem       bool      `json:"isSystem,omitempty" gorm:"type:boolean;not null;default:false"`
	CreatedAt      time.Time `json:"createdAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt      time.Time  `json:"-" gorm:"type:timestamptz;not null;default:now()"`
	DeletedAt      *time.Time `json:"-" gorm:"type:timestamptz"`
}

func (Stage) TableName() string { return "pipeline_stages" }
