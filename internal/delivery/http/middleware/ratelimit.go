package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
	"github.com/prayogopangestu/crm-system/backend/internal/domain"
)

// RateLimit returns a middleware that caps the number of requests per client
// IP within a sliding `window`. When the configured cache is nil the limiter
// fails open and lets every request through.
func RateLimit(cache domain.Cache, logger *slog.Logger, route string, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cache == nil {
				next.ServeHTTP(w, r)
				return
			}
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}
			bucket := time.Now().Unix() / int64(window.Seconds())
			key := "crm:rate:" + route + ":" + host + ":" + time.Unix(bucket*int64(window.Seconds()), 0).Format("20060102150405")
			allowed, err := cache.Allow(r.Context(), key, limit, window+time.Minute)
			if err != nil {
				logger.Warn("rate limiter unavailable", "error", err)
				next.ServeHTTP(w, r)
				return
			}
			if !allowed {
				response.Error(w, http.StatusTooManyRequests, "rate_limited", "terlalu banyak percobaan, coba lagi nanti", response.RequestID(r.Context()), nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
