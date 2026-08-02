package postgres

import "time"

type userModel struct {
	ID             string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string `gorm:"type:uuid"`
	FirstName      string
	LastName       string
	Email          string
	PasswordHash   *string
	GoogleID       *string `gorm:"type:text;uniqueIndex;column:google_id"`
	Role           string
	Status         string
	AvatarURL      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	RevokedAt      *time.Time
}

func (userModel) TableName() string { return "users" }
