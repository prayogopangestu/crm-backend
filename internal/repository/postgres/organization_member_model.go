package postgres

import "time"

type organizationMemberModel struct {
	ID             string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string     `gorm:"type:uuid;column:organization_id"`
	UserID         string     `gorm:"type:uuid;column:user_id"`
	Role           string
	Status         string
	InvitedBy      *string    `gorm:"type:uuid;column:invited_by"`
	JoinedAt       *time.Time `gorm:"type:timestamptz;column:joined_at"`
	RevokedAt      *time.Time `gorm:"type:timestamptz;column:revoked_at"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (organizationMemberModel) TableName() string { return "organization_members" }
