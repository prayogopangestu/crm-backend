package postgres

import "time"

type activityModel struct {
	ID             string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string `gorm:"type:uuid"`
	ActorID        string `gorm:"type:uuid"`
	ActorName      string
	Action         string
	Target         string
	IsHighlight    bool
	CreatedAt      time.Time
}

func (activityModel) TableName() string { return "activities" }
