package jwt

import (
	"testing"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
)

func TestManagerCreateAndParse(t *testing.T) {
	manager := New("01234567890123456789012345678901", time.Hour)
	raw, err := manager.Create("user-1", "org-1", domain.RoleAdmin, "Sarah")
	if err != nil {
		t.Fatal(err)
	}
	principal, err := manager.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if principal.UserID != "user-1" || principal.OrganizationID != "org-1" || principal.Role != domain.RoleAdmin {
		t.Fatalf("unexpected principal: %+v", principal)
	}
}

func TestManagerRejectsExpiredToken(t *testing.T) {
	manager := New("01234567890123456789012345678901", -time.Second)
	raw, err := manager.Create("user-1", "org-1", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Parse(raw); err == nil {
		t.Fatal("expected expired token to fail")
	}
}
