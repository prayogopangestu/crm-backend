package postgres

import "time"

type notificationModel struct {
	ID             string  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string  `gorm:"type:uuid"`
	UserID         *string `gorm:"type:uuid"`
	Title          string
	Message        string
	ReadAt         *time.Time
	CreatedAt      time.Time
}

func (notificationModel) TableName() string { return "notifications" }
