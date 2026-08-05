package contact

import (
	"context"
	"testing"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
)

type contactRepoStub struct {
	createCalled bool
}

func (r *contactRepoStub) List(context.Context, string, string, string, entities.ContactPage) (entities.ContactList, error) {
	return entities.ContactList{}, nil
}

func (r *contactRepoStub) Create(context.Context, domain.Principal, entities.Contact) (entities.Contact, error) {
	r.createCalled = true
	return entities.Contact{ID: "contact-1"}, nil
}

func (r *contactRepoStub) Update(context.Context, domain.Principal, string, entities.Contact) (entities.Contact, error) {
	return entities.Contact{}, nil
}

func (r *contactRepoStub) Delete(context.Context, domain.Principal, string) error {
	return nil
}

func TestViewerCanListButCannotCreateContact(t *testing.T) {
	repo := &contactRepoStub{}
	service := NewService(repo, support.CacheHelper{})
	viewer := domain.Principal{UserID: "user-1", OrganizationID: "org-1", Role: domain.RoleViewer}

	if _, err := service.List(context.Background(), viewer, "", "", 1, 20); err != nil {
		t.Fatalf("viewer list should be allowed: %v", err)
	}
	_, err := service.Create(context.Background(), viewer, Input{
		Name: "Sarah", Email: "sarah@example.com", Company: "Acme", Status: "Prospek Awal",
	})
	if err != domain.ErrForbidden {
		t.Fatalf("viewer create should be forbidden, got %v", err)
	}
	if repo.createCalled {
		t.Fatal("repository Create should not be called for viewer")
	}
}
