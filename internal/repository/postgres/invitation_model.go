package postgres

import "time"

type invitationModel struct {
	ID             string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string `gorm:"type:uuid"`
	UserID         string `gorm:"type:uuid"`
	Email          string
	Role           string
	TokenHash      string
	ExpiresAt      time.Time
	AcceptedAt     *time.Time
	CreatedAt      time.Time
}

func (invitationModel) TableName() string { return "user_invitations" }
