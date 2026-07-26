package postgres

import "time"

type contactModel struct {
	ID              string  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID  string  `gorm:"type:uuid"`
	OwnerID         *string `gorm:"type:uuid"`
	Name            string
	Email           string
	Company         string
	Role            string
	Status          string
	AvatarURL       string
	LastContactedAt time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

func (contactModel) TableName() string { return "contacts" }
