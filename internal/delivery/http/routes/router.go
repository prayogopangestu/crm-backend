package routes

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/middleware"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/repositories"
	"github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
	"github.com/prayogopangestu/crm-system/backend/pkg/encryption"
	"gorm.io/gorm"
)

// Deps bundles every dependency the application router needs to wire itself.
// Passing a single struct keeps the call sites (main.go and tests) stable as
// new features are added.
type Deps struct {
	DB          *gorm.DB
	Cache       domain.Cache
	CacheHelper support.CacheHelper
	Tokens      domain.TokenManager
	Logger      *slog.Logger
	Location    *time.Location
	BaseURL     string
	BcryptCost  int
	Origins     []string
	Cipher      *encryption.Cipher
	Sender      repositories.TelegramSender
	Ready       func() error
}

// NewRouter constructs the application chi router: global middleware, the
// health/readiness probes and every feature route mounted via WireAll.
func NewRouter(deps Deps) http.Handler {
	router := chi.NewRouter()
	registerGlobalMiddleware(router, deps.Logger, deps.Origins)
	mountHealth(router, deps.Logger, deps.Ready)

	WireAll(router, deps.DB, deps.Cache, deps.CacheHelper, deps.Tokens,
		deps.Logger, deps.Location, deps.BaseURL, deps.BcryptCost,
		IntegrationDeps{Cipher: deps.Cipher, Sender: deps.Sender, Logger: deps.Logger},
	)
	return router
}

// registerGlobalMiddleware attaches the cross-cutting middleware applied to
// every request: request id, real IP, panic recovery, access log and CORS.
func registerGlobalMiddleware(router chi.Router, logger *slog.Logger, origins []string) {
	router.Use(middleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(middleware.Recoverer(logger))
	router.Use(middleware.AccessLog(logger))
	router.Use(middleware.CORS(origins))
}

// mountHealth registers the liveness and readiness probes.
func mountHealth(router chi.Router, logger *slog.Logger, ready func() error) {
	router.Get("/healthz", health)
	router.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if ready != nil {
			if err := ready(); err != nil {
				logger.Warn("ready probe dependency check failed", "error", err)
			}
		}
		response.JSON(w, http.StatusOK, map[string]any{"status": "ready"})
	})
}

func health(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{"status": "ok"})
}
