package integration

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/repositories"
)

type TelegramInput struct {
	Enabled  bool
	BotToken string
	ChatID   string
}

type Cipher interface {
	Encrypt(plain string) (string, error)
	Decrypt(encoded string) (string, error)
}

type Logger interface {
	Error(msg string, args ...any)
}

type Service struct {
	repository repositories.IntegrationRepository
	cipher     Cipher
	sender     repositories.TelegramSender
}

func NewService(repository repositories.IntegrationRepository, cipher Cipher, sender repositories.TelegramSender) *Service {
	return &Service{repository: repository, cipher: cipher, sender: sender}
}

func (s *Service) Get(ctx context.Context, principal domain.Principal) (entities.Telegram, error) {
	if err := domain.RequireAdmin(principal); err != nil {
		return entities.Telegram{}, err
	}
	return s.repository.GetTelegram(ctx, principal.OrganizationID)
}

func (s *Service) Update(ctx context.Context, principal domain.Principal, input TelegramInput) (entities.Telegram, error) {
	if err := domain.RequireAdmin(principal); err != nil {
		return entities.Telegram{}, err
	}
	var encrypted string
	var err error
	if input.BotToken != "" {
		encrypted, err = s.cipher.Encrypt(input.BotToken)
		if err != nil {
			return entities.Telegram{}, err
		}
	}
	current, err := s.repository.GetTelegram(ctx, principal.OrganizationID)
	if err != nil {
		return entities.Telegram{}, err
	}
	if input.Enabled && input.BotToken == "" && !current.HasToken {
		return entities.Telegram{}, domain.ErrInvalidInput
	}
	if input.Enabled && input.ChatID == "" && current.ChatID == "" {
		return entities.Telegram{}, domain.ErrInvalidInput
	}
	if err := s.repository.UpsertTelegram(ctx, principal.OrganizationID, entities.Telegram{
		Enabled: input.Enabled, ChatID: input.ChatID, EncryptedToken: encrypted,
	}); err != nil {
		return entities.Telegram{}, err
	}
	return s.repository.GetTelegram(ctx, principal.OrganizationID)
}

func (s *Service) Test(ctx context.Context, principal domain.Principal) error {
	if err := domain.RequireAdmin(principal); err != nil {
		return err
	}
	value, err := s.repository.GetTelegram(ctx, principal.OrganizationID)
	if err != nil {
		return err
	}
	if !value.Enabled || !value.HasToken || value.ChatID == "" {
		return domain.ErrInvalidInput
	}
	token, err := s.cipher.Decrypt(value.EncryptedToken)
	if err != nil {
		return err
	}
	return s.sender.Send(ctx, token, value.ChatID, "CRM Enterprise: koneksi Telegram berhasil.")
}

type Worker struct {
	repository repositories.IntegrationRepository
	cipher     Cipher
	sender     repositories.TelegramSender
	logger     Logger
	interval   time.Duration
	batch      int
}

func NewWorker(repository repositories.IntegrationRepository, cipher Cipher, sender repositories.TelegramSender, logger Logger, interval time.Duration, batch int) *Worker {
	return &Worker{repository: repository, cipher: cipher, sender: sender, logger: logger, interval: interval, batch: batch}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		if err := w.process(ctx); err != nil && !errors.Is(err, context.Canceled) {
			w.logger.Error("outbox worker failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *Worker) process(ctx context.Context) error {
	events, err := w.repository.ClaimOutbox(ctx, w.batch)
	if err != nil {
		return err
	}
	for _, event := range events {
		if event.EventType != "telegram.deal_won" {
			_ = w.repository.CompleteOutbox(ctx, event.ID)
			continue
		}
		value, err := w.repository.GetTelegram(ctx, event.OrganizationID)
		if err == nil && (!value.Enabled || !value.HasToken || value.ChatID == "") {
			_ = w.repository.CompleteOutbox(ctx, event.ID)
			continue
		}
		var payload struct {
			Message string
		}
		if err == nil {
			err = json.Unmarshal(event.Payload, &payload)
		}
		var token string
		if err == nil {
			token, err = w.cipher.Decrypt(value.EncryptedToken)
		}
		if err == nil {
			err = w.sender.Send(ctx, token, value.ChatID, payload.Message)
		}
		if err == nil {
			_ = w.repository.CompleteOutbox(ctx, event.ID)
			continue
		}
		delay := time.Duration(1<<min(event.Attempts, 8)) * time.Minute
		_ = w.repository.RetryOutbox(ctx, event.ID, err.Error(), time.Now().Add(delay))
	}
	return nil
}
