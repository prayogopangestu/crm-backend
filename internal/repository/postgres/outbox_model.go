package postgres

import "time"

type outboxModel struct {
	ID             string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string `gorm:"type:uuid"`
	EventType      string
	Payload        []byte `gorm:"type:jsonb"`
	Attempts       int
	NextAttemptAt  time.Time
	ProcessedAt    *time.Time
	LastError      string
	CreatedAt      time.Time
}

func (outboxModel) TableName() string { return "outbox_events" }
