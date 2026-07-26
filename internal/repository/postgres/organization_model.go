package postgres

import "time"

type organizationModel struct {
	ID        string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (organizationModel) TableName() string { return "organizations" }
