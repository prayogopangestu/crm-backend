package entities

import "time"

type Contact struct {
	ID              string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID  string     `json:"-" gorm:"type:uuid;not null;column:organization_id"`
	OwnerID         string     `json:"-" gorm:"type:uuid;column:owner_id"`
	Name            string     `json:"name" gorm:"type:text;not null"`
	Email           string     `json:"email" gorm:"type:text;not null"`
	Company         string     `json:"company" gorm:"type:text;not null"`
	Role            string     `json:"role" gorm:"type:text;not null;default:''"`
	Status          string     `json:"status" gorm:"type:text;not null"`
	LastContacted   string     `json:"lastContacted" gorm:"-"`
	LastContactedAt time.Time  `json:"lastContactedAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
	Initials        string     `json:"initials" gorm:"-"`
	AvatarURL       string     `json:"avatarUrl,omitempty" gorm:"type:text;not null;default:''"`
	CreatedAt       time.Time  `json:"createdAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt       time.Time  `json:"updatedAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
	DeletedAt       *time.Time `json:"-" gorm:"type:timestamptz"`
}

func (Contact) TableName() string { return "contacts" }

type ContactPage struct {
	Page  int
	Limit int
}

type ContactList struct {
	Data  []Contact `json:"data"`
	Total int64     `json:"total"`
	Page  int       `json:"page"`
}
