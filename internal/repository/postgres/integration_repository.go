package postgres

import (
	"context"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/infrastructure/database/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IntegrationRepository struct {
	db       *gorm.DB
	location *time.Location
}

func NewIntegrationRepository(db *gorm.DB, location *time.Location) *IntegrationRepository {
	return &IntegrationRepository{db: db, location: location}
}

func (r *IntegrationRepository) GetTelegram(ctx context.Context, organizationID string) (entities.Telegram, error) {
	var record telegramModel
	err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).First(&record).Error
	if err != nil {
		if postgres.MapError(err) == domain.ErrNotFound {
			return entities.Telegram{OrganizationID: organizationID}, nil
		}
		return entities.Telegram{}, err
	}
	value := entities.Telegram{
		OrganizationID: record.OrganizationID, Enabled: record.Enabled, ChatID: record.ChatID,
		EncryptedToken: record.BotTokenEncrypted, UpdatedAt: record.UpdatedAt,
	}
	value.HasToken = value.EncryptedToken != ""
	if value.HasToken {
		value.WebhookURL = "https://api.telegram.org/bot***/sendMessage"
	}
	return value, nil
}

func (r *IntegrationRepository) UpsertTelegram(ctx context.Context, organizationID string, input entities.Telegram) error {
	record := telegramModel{
		OrganizationID: organizationID, Enabled: input.Enabled,
		ChatID: input.ChatID, BotTokenEncrypted: input.EncryptedToken,
	}
	return postgres.MapError(r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "organization_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"enabled":             gorm.Expr("EXCLUDED.enabled"),
			"chat_id":             gorm.Expr("CASE WHEN EXCLUDED.chat_id = '' THEN telegram_integrations.chat_id ELSE EXCLUDED.chat_id END"),
			"bot_token_encrypted": gorm.Expr("CASE WHEN EXCLUDED.bot_token_encrypted = '' THEN telegram_integrations.bot_token_encrypted ELSE EXCLUDED.bot_token_encrypted END"),
			"updated_at":          time.Now(),
		}),
	}).Create(&record).Error)
}

func (r *IntegrationRepository) ClaimOutbox(ctx context.Context, limit int) ([]entities.OutboxEvent, error) {
	events := make([]entities.OutboxEvent, 0)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var records []outboxModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("processed_at IS NULL AND next_attempt_at <= now()").
			Order("created_at").Limit(limit).Find(&records).Error; err != nil {
			return err
		}
		if len(records) == 0 {
			return nil
		}
		ids := make([]string, 0, len(records))
		for _, record := range records {
			ids = append(ids, record.ID)
			events = append(events, entities.OutboxEvent{
				ID: record.ID, OrganizationID: record.OrganizationID,
				EventType: record.EventType, Payload: record.Payload, Attempts: record.Attempts,
			})
		}
		return tx.Model(&outboxModel{}).Where("id IN ?", ids).
			UpdateColumn("next_attempt_at", gorm.Expr("now() + interval '1 minute'")).Error
	})
	return events, err
}

func (r *IntegrationRepository) CompleteOutbox(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&outboxModel{}).Where("id = ?", id).
		Updates(map[string]any{"processed_at": time.Now(), "last_error": ""}).Error
}

func (r *IntegrationRepository) RetryOutbox(ctx context.Context, id, reason string, next time.Time) error {
	return r.db.WithContext(ctx).Model(&outboxModel{}).Where("id = ?", id).
		Updates(map[string]any{
			"attempts": gorm.Expr("attempts + 1"), "last_error": reason, "next_attempt_at": next,
		}).Error
}
