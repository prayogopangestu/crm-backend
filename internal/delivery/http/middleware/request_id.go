package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
)

// RequestID propagates an X-Request-ID header through the request context.
// When the inbound request has no header a fresh UUID is generated so every
// request carries a stable correlation id.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Header.Get("X-Request-ID")
		if value == "" {
			value = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", value)
		next.ServeHTTP(w, response.WithRequestID(r, value))
	})
}
