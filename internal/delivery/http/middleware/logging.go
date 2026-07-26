package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
)

// AccessLog returns a middleware that records each HTTP request as a single
// structured log line including method, path, duration and request id.
func AccessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Info("http request",
				"method", r.Method, "path", r.URL.Path,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", response.RequestID(r.Context()),
			)
		})
	}
}
