package routes

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/handler"
	pgrepo "github.com/prayogopangestu/crm-system/backend/internal/repository/postgres"
	pipelineusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/pipeline"
	usecasesupport "github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
	"gorm.io/gorm"
)

// PipelineRoutes wires the pipeline feature and registers the pipeline stage
// endpoints on the router.
func PipelineRoutes(router chi.Router, db *gorm.DB, location *time.Location, cacheHelper usecasesupport.CacheHelper, logger *slog.Logger) {
	repo := pgrepo.NewPipelineRepository(db, location)
	service := pipelineusecase.NewService(repo, cacheHelper)
	handler.NewPipelineHandler(service, logger).Routes(router)
}
