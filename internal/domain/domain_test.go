package domain

import "testing"

func TestWorkspaceRolePermissions(t *testing.T) {
	viewer := Principal{Role: RoleViewer}
	if err := RequireCanReadCRM(viewer); err != nil {
		t.Fatalf("viewer should read CRM: %v", err)
	}
	if err := RequireCanWriteCRM(viewer); err != ErrForbidden {
		t.Fatalf("viewer write should be forbidden, got %v", err)
	}
	if err := RequireWorkspaceAdmin(viewer); err != ErrForbidden {
		t.Fatalf("viewer admin should be forbidden, got %v", err)
	}

	owner := Principal{Role: RoleOwner}
	if err := RequireCanWriteCRM(owner); err != nil {
		t.Fatalf("owner should write CRM: %v", err)
	}
	if err := RequireWorkspaceAdmin(owner); err != nil {
		t.Fatalf("owner should manage workspace: %v", err)
	}
}

func TestInvitableRoles(t *testing.T) {
	for _, role := range []string{RoleAdmin, RoleSales, RoleViewer} {
		if !InvitableRole(role) {
			t.Fatalf("%s should be invitable", role)
		}
	}
	if InvitableRole(RoleOwner) {
		t.Fatal("owner should not be invitable")
	}
}
