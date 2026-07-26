package repositories

import (
	"context"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
)

type IntegrationRepository interface {
	GetTelegram(ctx context.Context, organizationID string) (entities.Telegram, error)
	UpsertTelegram(ctx context.Context, organizationID string, input entities.Telegram) error
	ClaimOutbox(ctx context.Context, limit int) ([]entities.OutboxEvent, error)
	CompleteOutbox(ctx context.Context, id string) error
	RetryOutbox(ctx context.Context, id, reason string, next time.Time) error
}

type TelegramSender interface {
	Send(ctx context.Context, token, chatID, message string) error
}
