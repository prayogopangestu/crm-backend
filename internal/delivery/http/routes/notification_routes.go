package routes

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/handler"
	pgrepo "github.com/prayogopangestu/crm-system/backend/internal/repository/postgres"
	notificationusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/notification"
	"gorm.io/gorm"
)

// NotificationRoutes wires the notification feature and registers the
// notification endpoints on the router.
func NotificationRoutes(router chi.Router, db *gorm.DB, location *time.Location, logger *slog.Logger) {
	repo := pgrepo.NewNotificationRepository(db, location)
	service := notificationusecase.NewService(repo)
	handler.NewNotificationHandler(service, logger).Routes(router)
}
