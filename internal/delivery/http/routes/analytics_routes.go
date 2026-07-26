package routes

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/handler"
	pgrepo "github.com/prayogopangestu/crm-system/backend/internal/repository/postgres"
	analyticsusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/analytics"
	usecasesupport "github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
	"gorm.io/gorm"
)

// AnalyticsRoutes wires the analytics feature and registers the dashboard /
// reports endpoints on the router.
func AnalyticsRoutes(router chi.Router, db *gorm.DB, location *time.Location, cacheHelper usecasesupport.CacheHelper, logger *slog.Logger) {
	repo := pgrepo.NewAnalyticsRepository(db, location)
	service := analyticsusecase.NewService(repo, cacheHelper, location)
	handler.NewAnalyticsHandler(service, logger).Routes(router)
}
