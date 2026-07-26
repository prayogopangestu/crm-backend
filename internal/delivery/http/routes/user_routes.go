package routes

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/handler"
	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	pgrepo "github.com/prayogopangestu/crm-system/backend/internal/repository/postgres"
	userusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/user"
	usecasesupport "github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
	"gorm.io/gorm"
)

// UserDeps bundles the configuration values required by the user usecase.
type UserDeps struct {
	Tokens      domain.TokenManager
	CacheHelper usecasesupport.CacheHelper
	BaseURL     string
	BcryptCost  int
	Logger      *slog.Logger
}

// registerUserRoutes wires the user feature and registers both the public
// (auth) routes and the protected profile/team routes. The authLimiter is
// applied only to the public auth endpoints.
func registerUserRoutes(
	publicRouter chi.Router,
	protectedRouter chi.Router,
	db *gorm.DB,
	location *time.Location,
	deps UserDeps,
	authLimiter func(http.Handler) http.Handler,
) {
	repo := pgrepo.NewUserRepository(db, location)
	service := userusecase.NewService(repo, deps.CacheHelper, deps.Tokens, deps.BaseURL, deps.BcryptCost)
	h := handler.NewUserHandler(service, deps.Logger)
	h.PublicRoutes(publicRouter, authLimiter)
	h.ProtectedRoutes(protectedRouter)
}
