package entities

import "time"

type Organization struct {
	ID        string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    `json:"name" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"createdAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt time.Time  `json:"updatedAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
	DeletedAt *time.Time `json:"-" gorm:"type:timestamptz"`
}

func (Organization) TableName() string { return "organizations" }
