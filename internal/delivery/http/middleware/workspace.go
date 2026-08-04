package middleware

import (
	"net/http"
	"strings"

	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"gorm.io/gorm"
)

// WorkspaceAccess verifies that the authenticated principal is still an active
// member of the selected workspace. If X-Workspace-ID is present, it becomes
// the active workspace for the request after membership validation.
func WorkspaceAccess(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := domain.PrincipalFromContext(r.Context())
			if !ok || principal.UserID == "" {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "Bearer token diperlukan", response.RequestID(r.Context()), nil)
				return
			}

			workspaceID := strings.TrimSpace(r.Header.Get("X-Workspace-ID"))
			if workspaceID == "" {
				workspaceID = principal.OrganizationID
			}
			if workspaceID == "" {
				response.Error(w, http.StatusForbidden, "forbidden", "Anda tidak memiliki akses", response.RequestID(r.Context()), nil)
				return
			}

			var row struct {
				Role string
				Name string
			}
			err := db.WithContext(r.Context()).Table("organization_members AS om").
				Select("om.role, trim(u.first_name || ' ' || u.last_name) AS name").
				Joins("JOIN users u ON u.id = om.user_id").
				Where("om.user_id = ? AND om.organization_id = ? AND om.status = 'Aktif' AND om.revoked_at IS NULL AND u.revoked_at IS NULL",
					principal.UserID, workspaceID).
				First(&row).Error
			if err != nil {
				response.Error(w, http.StatusForbidden, "forbidden", "Anda tidak memiliki akses", response.RequestID(r.Context()), nil)
				return
			}

			principal.OrganizationID = workspaceID
			principal.Role = row.Role
			if row.Name != "" {
				principal.Name = row.Name
			}
			next.ServeHTTP(w, r.WithContext(domain.WithPrincipal(r.Context(), principal)))
		})
	}
}
