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

// OrganizationMember is the many-to-many link between a user and a workspace
// (organization). It is the source of truth for which workspaces a user can
// access and with which role. A user can hold memberships in many workspaces.
type OrganizationMember struct {
	ID             string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string     `json:"organizationId" gorm:"type:uuid;not null;column:organization_id"`
	UserID         string     `json:"userId" gorm:"type:uuid;not null;column:user_id"`
	Role           string     `json:"role" gorm:"type:text;not null"`
	Status         string     `json:"status" gorm:"type:text;not null;default:'Aktif'"`
	InvitedBy      *string    `json:"-" gorm:"type:uuid;column:invited_by"`
	JoinedAt       *time.Time `json:"joinedAt,omitempty" gorm:"type:timestamptz;column:joined_at"`
	RevokedAt      *time.Time `json:"-" gorm:"type:timestamptz;column:revoked_at"`
	CreatedAt      time.Time  `json:"createdAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt      time.Time  `json:"updatedAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
}

func (OrganizationMember) TableName() string { return "organization_members" }

// WorkspaceSummary is the lightweight projection returned to clients listing
// the workspaces a user may access.
type WorkspaceSummary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	IsDefault bool   `json:"isDefault"`
}
