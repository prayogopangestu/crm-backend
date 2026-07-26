package middleware

import (
	"net/http"
	"strings"

	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
	"github.com/prayogopangestu/crm-system/backend/internal/domain"
)

// Authenticate returns a middleware that validates the Bearer JWT in the
// Authorization header and stores the resulting Principal in the request
// context for downstream handlers.
func Authenticate(tokens domain.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "Bearer token diperlukan", response.RequestID(r.Context()), nil)
				return
			}
			principal, err := tokens.Parse(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "token tidak valid atau kedaluwarsa", response.RequestID(r.Context()), nil)
				return
			}
			next.ServeHTTP(w, r.WithContext(domain.WithPrincipal(r.Context(), principal)))
		})
	}
}
