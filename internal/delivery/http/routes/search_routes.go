package routes

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/handler"
	pgrepo "github.com/prayogopangestu/crm-system/backend/internal/repository/postgres"
	searchusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/search"
	usecasesupport "github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
	"gorm.io/gorm"
)

// SearchRoutes wires the global search feature and registers the search
// endpoint on the router.
func SearchRoutes(router chi.Router, db *gorm.DB, location *time.Location, cacheHelper usecasesupport.CacheHelper, logger *slog.Logger) {
	repo := pgrepo.NewSearchRepository(db, location)
	service := searchusecase.NewService(repo, cacheHelper)
	handler.NewSearchHandler(service, logger).Routes(router)
}
