package middleware

import (
	"log/slog"
	"net/http"

	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
)

// Recoverer returns a middleware that converts panics into 500 responses so a
// single failing handler cannot crash the whole process.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if value := recover(); value != nil {
					logger.Error("panic recovered", "panic", value, "request_id", response.RequestID(r.Context()))
					response.Error(w, http.StatusInternalServerError, "internal_error", "terjadi kesalahan internal", response.RequestID(r.Context()), nil)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
