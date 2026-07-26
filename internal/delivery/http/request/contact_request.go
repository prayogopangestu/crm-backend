package request

import (
	contactusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/contact"
)

// ContactCreate represents the JSON body for create/update contact endpoints.
type ContactCreate struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Company   string `json:"company"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	AvatarURL string `json:"avatarUrl"`
}

// ToInput converts the request DTO into the contact usecase Input.
func (r ContactCreate) ToInput() contactusecase.Input {
	return contactusecase.Input{
		Name: r.Name, Email: r.Email, Company: r.Company,
		Role: r.Role, Status: r.Status, AvatarURL: r.AvatarURL,
	}
}
