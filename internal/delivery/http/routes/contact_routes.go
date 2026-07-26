package routes

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/handler"
	pgrepo "github.com/prayogopangestu/crm-system/backend/internal/repository/postgres"
	contactusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/contact"
	usecasesupport "github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
	"gorm.io/gorm"
)

// ContactRoutes wires the contact feature (repository -> service -> handler) and
// registers the contact endpoints on the provided router.
func ContactRoutes(router chi.Router, db *gorm.DB, location *time.Location, cacheHelper usecasesupport.CacheHelper, logger *slog.Logger) {
	repo := pgrepo.NewContactRepository(db, location)
	service := contactusecase.NewService(repo, cacheHelper)
	handler.NewContactHandler(service, logger).Routes(router)
}
