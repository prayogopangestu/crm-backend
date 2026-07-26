package routes

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/handler"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/repositories"
	pgrepo "github.com/prayogopangestu/crm-system/backend/internal/repository/postgres"
	integrationusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/integration"
	"gorm.io/gorm"
)

// IntegrationDeps bundles the external dependencies required by the
// integration usecase (cipher for token encryption and a telegram sender).
type IntegrationDeps struct {
	Cipher integrationusecase.Cipher
	Sender repositories.TelegramSender
	Logger *slog.Logger
}

// IntegrationRoutes wires the integration feature and registers the telegram
// integration endpoints on the router.
func IntegrationRoutes(router chi.Router, db *gorm.DB, location *time.Location, deps IntegrationDeps) {
	repo := pgrepo.NewIntegrationRepository(db, location)
	service := integrationusecase.NewService(repo, deps.Cipher, deps.Sender)
	handler.NewIntegrationHandler(service, deps.Logger).Routes(router)
}
