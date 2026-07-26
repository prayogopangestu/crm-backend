package routes

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/handler"
	pgrepo "github.com/prayogopangestu/crm-system/backend/internal/repository/postgres"
	dealusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/deal"
	usecasesupport "github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
	"gorm.io/gorm"
)

// DealRoutes wires the deal feature and registers the deal endpoints on the router.
func DealRoutes(router chi.Router, db *gorm.DB, location *time.Location, cacheHelper usecasesupport.CacheHelper, logger *slog.Logger) {
	repo := pgrepo.NewDealRepository(db, location)
	service := dealusecase.NewService(repo, cacheHelper)
	handler.NewDealHandler(service, logger).Routes(router)
}
