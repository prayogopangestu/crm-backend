package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"github.com/prayogopangestu/crm-system/backend/internal/infrastructure/database/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository struct {
	db       *gorm.DB
	location *time.Location
}

func NewUserRepository(db *gorm.DB, location *time.Location) *UserRepository {
	return &UserRepository{db: db, location: location}
}

func (r *UserRepository) Register(ctx context.Context, orgName string, value entities.User) (entities.User, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		organization := organizationModel{Name: orgName}
		if err := tx.Create(&organization).Error; err != nil {
			return postgres.MapError(err)
		}
		passwordHash := value.PasswordHash
		record := userModel{
			OrganizationID: organization.ID, FirstName: value.FirstName, LastName: value.LastName,
			Email: strings.ToLower(value.Email), PasswordHash: &passwordHash, Role: value.Role, Status: "Aktif",
		}
		if err := tx.Create(&record).Error; err != nil {
			return postgres.MapError(err)
		}
		value = toUserEntity(record)
		stages := []stageModel{
			{OrganizationID: organization.ID, Key: "lead", Name: "Lead Masuk", Color: "bg-primary-container", Position: 1, IsSystem: true},
			{OrganizationID: organization.ID, Key: "contacted", Name: "Dihubungi", Color: "bg-secondary-container", Position: 2, IsSystem: true},
			{OrganizationID: organization.ID, Key: "meeting", Name: "Meeting", Color: "bg-tertiary-container", Position: 3, IsSystem: true},
			{OrganizationID: organization.ID, Key: "negotiation", Name: "Negosiasi", Color: "bg-primary-fixed", Position: 4, IsSystem: true},
			{OrganizationID: organization.ID, Key: "won", Name: "Deal Won", Color: "bg-surface-tint", Position: 5, IsSystem: true},
			{OrganizationID: organization.ID, Key: "lost", Name: "Deal Lost", Color: "bg-error-container", Position: 6, IsSystem: true},
		}
		if err := tx.Create(&stages).Error; err != nil {
			return postgres.MapError(err)
		}
		now := time.Now().In(r.location)
		goals := make([]goalModel, 0, 3)
		for offset, target := range []int64{1_000_000_000, 900_000_000, 900_000_000} {
			goals = append(goals, goalModel{
				OrganizationID: organization.ID,
				Month:          time.Date(now.Year(), now.Month()-time.Month(offset), 1, 0, 0, 0, 0, r.location),
				Goal:           target,
			})
		}
		return postgres.MapError(tx.Create(&goals).Error)
	})
	return value, err
}

func (r *UserRepository) ByEmail(ctx context.Context, email string) (entities.User, error) {
	var record userModel
	err := r.db.WithContext(ctx).Where("lower(email) = lower(?) AND revoked_at IS NULL", email).First(&record).Error
	if err != nil {
		return entities.User{}, postgres.MapError(err)
	}
	return toUserEntity(record), nil
}

func (r *UserRepository) ByID(ctx context.Context, organizationID, userID string) (entities.User, error) {
	var record userModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ? AND revoked_at IS NULL", userID, organizationID).
		First(&record).Error
	if err != nil {
		return entities.User{}, postgres.MapError(err)
	}
	return toUserEntity(record), nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, principal domain.Principal, firstName, lastName, email string) (entities.User, error) {
	result := r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ? AND organization_id = ? AND revoked_at IS NULL", principal.UserID, principal.OrganizationID).
		Updates(map[string]any{
			"first_name": firstName, "last_name": lastName,
			"email": strings.ToLower(email), "updated_at": time.Now(),
		})
	if result.Error != nil {
		return entities.User{}, postgres.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return entities.User{}, domain.ErrNotFound
	}
	return r.ByID(ctx, principal.OrganizationID, principal.UserID)
}

func (r *UserRepository) AcceptInvite(ctx context.Context, tokenHash, passwordHash string) (entities.User, error) {
	var invitation invitationModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ?", tokenHash).First(&invitation).Error; err != nil {
			return postgres.MapError(err)
		}
		if invitation.AcceptedAt != nil {
			return domain.ErrInviteUsed
		}
		if time.Now().After(invitation.ExpiresAt) {
			return domain.ErrInviteExpired
		}
		result := tx.Model(&userModel{}).
			Where("id = ? AND organization_id = ? AND revoked_at IS NULL", invitation.UserID, invitation.OrganizationID).
			Updates(map[string]any{"password_hash": passwordHash, "status": "Aktif", "updated_at": time.Now()})
		if result.Error != nil {
			return postgres.MapError(result.Error)
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		return postgres.MapError(tx.Model(&invitationModel{}).
			Where("id = ?", invitation.ID).Update("accepted_at", time.Now()).Error)
	})
	if err != nil {
		return entities.User{}, err
	}
	return r.ByID(ctx, invitation.OrganizationID, invitation.UserID)
}

func (r *UserRepository) ListTeam(ctx context.Context, organizationID string) ([]entities.User, error) {
	var records []userModel
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND revoked_at IS NULL", organizationID).
		Order("created_at").Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]entities.User, 0, len(records))
	for _, record := range records {
		items = append(items, toUserEntity(record))
	}
	return items, nil
}

func (r *UserRepository) InviteMember(ctx context.Context, principal domain.Principal, value entities.User, invitation entities.Invitation) (entities.User, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		record := userModel{
			OrganizationID: principal.OrganizationID, FirstName: value.FirstName, LastName: value.LastName,
			Email: strings.ToLower(value.Email), Role: value.Role, Status: "Menunggu",
		}
		if err := tx.Create(&record).Error; err != nil {
			return postgres.MapError(err)
		}
		value = toUserEntity(record)
		if err := tx.Create(&invitationModel{
			OrganizationID: principal.OrganizationID, UserID: record.ID,
			Email: strings.ToLower(value.Email), Role: value.Role,
			TokenHash: invitation.TokenHash, ExpiresAt: invitation.ExpiresAt,
		}).Error; err != nil {
			return postgres.MapError(err)
		}
		return postgres.MapError(tx.Create(&activityModel{
			OrganizationID: principal.OrganizationID, ActorID: principal.UserID,
			ActorName: principal.Name, Action: "mengundang anggota tim", Target: value.Email,
		}).Error)
	})
	return value, err
}

func (r *UserRepository) RevokeMember(ctx context.Context, organizationID, userID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ? AND organization_id = ? AND revoked_at IS NULL", userID, organizationID).
		Updates(map[string]any{"status": "Dicabut", "revoked_at": now, "updated_at": now})
	if result.Error != nil {
		return postgres.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func toUserEntity(record userModel) entities.User {
	passwordHash := ""
	if record.PasswordHash != nil {
		passwordHash = *record.PasswordHash
	}
	value := entities.User{
		ID: record.ID, OrganizationID: record.OrganizationID,
		FirstName: record.FirstName, LastName: record.LastName, Email: record.Email,
		PasswordHash: passwordHash, Role: record.Role, Status: record.Status,
		AvatarURL: record.AvatarURL, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}
	value.Name = strings.TrimSpace(value.FirstName + " " + value.LastName)
	value.Initials = postgres.Initials(value.Name)
	return value
}
