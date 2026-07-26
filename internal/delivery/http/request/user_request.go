package request

import (
	"strings"

	userusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/user"
)

// UserRegister represents the JSON body for the register endpoint.
type UserRegister struct {
	Name        string `json:"name"`
	FullName    string `json:"fullName"`
	CompanyName string `json:"companyName"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

// ToInput converts the request DTO into the user usecase RegisterInput.
// Falls back to FullName when Name is empty so the API accepts both fields.
func (r UserRegister) ToInput() userusecase.RegisterInput {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		name = strings.TrimSpace(r.FullName)
	}
	return userusecase.RegisterInput{
		Name: name, CompanyName: r.CompanyName, Email: r.Email, Password: r.Password,
	}
}

// UserLogin represents the JSON body for the login endpoint.
type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ToInput converts the request DTO into the user usecase LoginInput.
func (r UserLogin) ToInput() userusecase.LoginInput {
	return userusecase.LoginInput{Email: r.Email, Password: r.Password}
}

// AcceptInvite represents the JSON body for the accept-invite endpoint.
type AcceptInvite struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// UpdateProfile represents the JSON body for the update profile endpoint.
type UpdateProfile struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

// ToInput converts the request DTO into the user usecase UpdateProfileInput.
func (r UpdateProfile) ToInput() userusecase.UpdateProfileInput {
	return userusecase.UpdateProfileInput{FirstName: r.FirstName, LastName: r.LastName, Email: r.Email}
}

// InviteMember represents the JSON body for the team invite endpoint.
type InviteMember struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// ToInput converts the request DTO into the user usecase InviteInput.
func (r InviteMember) ToInput() userusecase.InviteInput {
	return userusecase.InviteInput{Name: r.Name, Email: r.Email, Role: r.Role}
}
