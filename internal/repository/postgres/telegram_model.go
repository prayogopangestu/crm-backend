package postgres

import "time"

type telegramModel struct {
	OrganizationID    string `gorm:"type:uuid;primaryKey"`
	BotTokenEncrypted string
	ChatID            string
	Enabled           bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (telegramModel) TableName() string { return "telegram_integrations" }
