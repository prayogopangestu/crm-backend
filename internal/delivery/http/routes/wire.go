package routes

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/middleware"
	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	usecasesupport "github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
	"gorm.io/gorm"
)

// WireAll registers every feature route on the router. Public auth routes are
// mounted on the root router (with an auth rate-limiter); protected feature
// routes are mounted inside an authenticated group.
//
// When `db` is nil wiring is skipped (used by tests that only exercise
// global middleware). This is the single entry point called from
// cmd/api/main.go to wire the delivery layer; per-feature wiring lives in
// <feature>_routes.go files.
func WireAll(
	router chi.Router,
	db *gorm.DB,
	cache domain.Cache,
	cacheHelper usecasesupport.CacheHelper,
	tokens domain.TokenManager,
	logger *slog.Logger,
	location *time.Location,
	baseURL string,
	bcryptCost int,
	cipherDeps IntegrationDeps,
	googleCfg GoogleConfig,
) {
	if db == nil {
		return
	}
	authLimiter := middleware.RateLimit(cache, logger, "auth", 5, time.Minute)

	router.Group(func(protected chi.Router) {
		protected.Use(middleware.Authenticate(tokens))

		registerUserRoutes(router, protected, db, location, UserDeps{
			Tokens: tokens, CacheHelper: cacheHelper, BaseURL: baseURL,
			BcryptCost: bcryptCost, Logger: logger,
			Google: googleCfg, Cache: cache,
		}, authLimiter)

		ContactRoutes(protected, db, location, cacheHelper, logger)
		DealRoutes(protected, db, location, cacheHelper, logger)
		TaskRoutes(protected, db, location, cacheHelper, logger)
		PipelineRoutes(protected, db, location, cacheHelper, logger)
		AnalyticsRoutes(protected, db, location, cacheHelper, logger)
		IntegrationRoutes(protected, db, location, cipherDeps)
		NotificationRoutes(protected, db, location, logger)
		SearchRoutes(protected, db, location, cacheHelper, logger)
	})
}
