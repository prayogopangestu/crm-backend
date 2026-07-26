package request

import (
	integrationusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/integration"
)

// TelegramUpdate represents the JSON body for the update telegram integration endpoint.
type TelegramUpdate struct {
	Enabled  bool   `json:"enabled"`
	BotToken string `json:"botToken"`
	ChatID   string `json:"chatId"`
}

// ToInput converts the request DTO into the integration usecase TelegramInput.
func (r TelegramUpdate) ToInput() integrationusecase.TelegramInput {
	return integrationusecase.TelegramInput{
		Enabled: r.Enabled, BotToken: r.BotToken, ChatID: r.ChatID,
	}
}
