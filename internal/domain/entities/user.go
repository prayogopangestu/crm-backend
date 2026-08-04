package entities

import "time"

type User struct {
	ID             string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string     `json:"-" gorm:"type:uuid;not null;column:organization_id"`
	FirstName      string     `json:"firstName" gorm:"type:text;not null"`
	LastName       string     `json:"lastName" gorm:"type:text;not null;default:''"`
	Name           string     `json:"name,omitempty" gorm:"-"`
	Email          string     `json:"email" gorm:"type:text;not null"`
	PasswordHash   string     `json:"-" gorm:"type:text"`
	GoogleID       *string    `json:"-" gorm:"type:text;uniqueIndex;column:google_id"`
	Role           string     `json:"role" gorm:"type:text;not null"`
	Status         string     `json:"status,omitempty" gorm:"type:text;not null;default:'Aktif'"`
	AvatarURL      string     `json:"avatarUrl" gorm:"type:text;not null;default:''"`
	Initials       string     `json:"initials,omitempty" gorm:"-"`
	CreatedAt      time.Time  `json:"createdAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt      time.Time  `json:"updatedAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
	RevokedAt      *time.Time `json:"-" gorm:"type:timestamptz"`
	DeletedAt      *time.Time `json:"-" gorm:"type:timestamptz"`
}

func (User) TableName() string { return "users" }

type Invitation struct {
	ID             string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string     `json:"-" gorm:"type:uuid;not null;column:organization_id"`
	UserID         *string    `json:"-" gorm:"type:uuid;column:user_id"`
	Email          string     `json:"email" gorm:"type:text;not null"`
	Role           string     `json:"role" gorm:"type:text;not null"`
	TokenHash      string     `json:"-" gorm:"type:text;not null;uniqueIndex"`
	ExpiresAt      time.Time  `json:"expiresAt" gorm:"type:timestamptz;not null"`
	AcceptedAt     *time.Time `json:"-" gorm:"type:timestamptz"`
	InvitedBy      *string    `json:"-" gorm:"type:uuid;column:invited_by"`
	CreatedAt      time.Time  `json:"createdAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
	DeletedAt      *time.Time `json:"-" gorm:"type:timestamptz"`
}

func (Invitation) TableName() string { return "user_invitations" }
