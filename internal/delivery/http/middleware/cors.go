package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
)

// CORS returns a middleware that applies the Cross-Origin Resource Sharing
// policy. Only origins listed in `origins` are allowed; "*" permits any
// origin. Preflight OPTIONS requests are answered directly.
func CORS(origins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(origins))
	allowAll := false
	for _, origin := range origins {
		origin = normalizeOrigin(origin)
		if origin == "" {
			continue
		}
		if origin == "*" {
			allowAll = true
			continue
		}
		allowed[origin] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if originAllowed(origin, allowed, allowAll) {
				setCORSHeaders(w, origin, r.Header.Get("Access-Control-Request-Headers"))
			}
			if r.Method == http.MethodOptions {
				if origin != "" && !originAllowed(origin, allowed, allowAll) {
					response.Error(w, http.StatusForbidden, "cors_origin_forbidden", "origin tidak diizinkan", response.RequestID(r.Context()), nil)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func normalizeOrigin(origin string) string {
	origin = strings.TrimSpace(origin)
	if origin == "" || origin == "*" {
		return origin
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return strings.TrimRight(origin, "/")
	}
	return parsed.Scheme + "://" + parsed.Host
}

func originAllowed(origin string, allowed map[string]bool, allowAll bool) bool {
	if origin == "" {
		return false
	}
	return allowAll || allowed[normalizeOrigin(origin)]
}

func setCORSHeaders(w http.ResponseWriter, origin, requestedHeaders string) {
	headers := w.Header()
	headers.Set("Access-Control-Allow-Origin", origin)
	headers.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	if requestedHeaders != "" {
		headers.Set("Access-Control-Allow-Headers", requestedHeaders)
	} else {
		headers.Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")
	}
	headers.Set("Access-Control-Expose-Headers", "X-Request-ID")
	headers.Set("Access-Control-Max-Age", "600")
	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")
}
