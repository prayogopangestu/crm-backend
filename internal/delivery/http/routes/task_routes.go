package routes

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/handler"
	pgrepo "github.com/prayogopangestu/crm-system/backend/internal/repository/postgres"
	taskusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/task"
	usecasesupport "github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
	"gorm.io/gorm"
)

// TaskRoutes wires the task feature and registers the task endpoints on the router.
// The location is used for date/time parsing inside the task service.
func TaskRoutes(router chi.Router, db *gorm.DB, location *time.Location, cacheHelper usecasesupport.CacheHelper, logger *slog.Logger) {
	repo := pgrepo.NewTaskRepository(db, location)
	service := taskusecase.NewService(repo, cacheHelper, location)
	handler.NewTaskHandler(service, logger).Routes(router)
}
