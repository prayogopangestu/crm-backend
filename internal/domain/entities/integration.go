package entities

import "time"

type Telegram struct {
	OrganizationID    string    `json:"-" gorm:"type:uuid;primaryKey"`
	Enabled           bool      `json:"enabled" gorm:"type:boolean;not null;default:false"`
	WebhookURL        string    `json:"webhookUrl" gorm:"-"`
	ChatID            string    `json:"chatId,omitempty" gorm:"type:text;not null;default:''"`
	HasToken          bool      `json:"hasToken" gorm:"-"`
	EncryptedToken    string    `json:"-" gorm:"type:text;not null;default:'';column:bot_token_encrypted"`
	CreatedAt         time.Time `json:"-" gorm:"type:timestamptz;not null;default:now()"`
	DeletedAt         *time.Time `json:"-" gorm:"type:timestamptz"`
	UpdatedAt         time.Time  `json:"updatedAt,omitempty" gorm:"type:timestamptz;not null;default:now()"`
}

func (Telegram) TableName() string { return "telegram_integrations" }

type OutboxEvent struct {
	ID             string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string     `gorm:"type:uuid;not null;column:organization_id"`
	EventType      string     `gorm:"type:text;not null"`
	Payload        []byte     `gorm:"type:jsonb;not null"`
	Attempts       int        `gorm:"type:integer;not null;default:0"`
	NextAttemptAt  time.Time  `gorm:"type:timestamptz;not null;default:now()"`
	ProcessedAt    *time.Time `gorm:"type:timestamptz"`
	LastError      string     `gorm:"type:text;not null;default:''"`
	DeletedAt      *time.Time `gorm:"type:timestamptz"`
	CreatedAt      time.Time  `gorm:"type:timestamptz;not null;default:now()"`
}

func (OutboxEvent) TableName() string { return "outbox_events" }
