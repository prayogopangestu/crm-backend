package entities

import "time"

type Notification struct {
	ID             string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string     `json:"-" gorm:"type:uuid;not null;column:organization_id"`
	UserID         string     `json:"-" gorm:"type:uuid;column:user_id"`
	Title          string     `json:"title" gorm:"type:text;not null"`
	Message        string     `json:"message" gorm:"type:text;not null"`
	ReadAt         *time.Time `json:"-" gorm:"type:timestamptz"`
	Time           string     `json:"time" gorm:"-"`
	Read           bool       `json:"read" gorm:"-"`
	DeletedAt      *time.Time `json:"-" gorm:"type:timestamptz"`
	CreatedAt      time.Time  `json:"createdAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
}

func (Notification) TableName() string { return "notifications" }
