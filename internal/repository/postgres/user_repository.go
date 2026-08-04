package postgres

import (
	"context"
	"errors"
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
	return r.createWithOrganization(ctx, orgName, value, true)
}

func (r *UserRepository) CreateGoogleUser(ctx context.Context, orgName string, value entities.User) (entities.User, error) {
	return r.createWithOrganization(ctx, orgName, value, false)
}

// createWithOrganization provisions a new workspace, a global user as its
// Owner, the Owner membership and seeds pipeline stages + performance goals.
// users.organization_id/role are still populated for backwards compatibility.
func (r *UserRepository) createWithOrganization(ctx context.Context, orgName string, value entities.User, withPassword bool) (entities.User, error) {
	role := value.Role
	if role == "" {
		role = domain.RoleOwner
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		organization := organizationModel{Name: orgName}
		if err := tx.Create(&organization).Error; err != nil {
			return postgres.MapError(err)
		}
		passwordHash := value.PasswordHash
		record := userModel{
			OrganizationID: organization.ID, FirstName: value.FirstName, LastName: value.LastName,
			Email: strings.ToLower(value.Email), GoogleID: value.GoogleID,
			AvatarURL: value.AvatarURL, Role: role, Status: "Aktif",
		}
		if withPassword {
			record.PasswordHash = &passwordHash
		}
		if err := tx.Create(&record).Error; err != nil {
			return postgres.MapError(err)
		}
		if err := createOwnerMembership(tx, organization.ID, record.ID); err != nil {
			return postgres.MapError(err)
		}
		value = toUserEntity(record)
		return seedOrganization(tx, organization.ID, r.location)
	})
	if err != nil {
		return entities.User{}, err
	}
	return value, err
}

func (r *UserRepository) ByEmail(ctx context.Context, email string) (entities.User, error) {
	return r.ByEmailGlobal(ctx, email)
}

// ByEmailGlobal looks up a user identity by email across all workspaces.
// A user is now a global identity; the first non-revoked match wins. Historical
// duplicate emails (from the legacy invite flow) are ordered by created_at so
// the original account is preferred.
func (r *UserRepository) ByEmailGlobal(ctx context.Context, email string) (entities.User, error) {
	var record userModel
	err := r.db.WithContext(ctx).
		Where("lower(email) = lower(?) AND revoked_at IS NULL", email).
		Order("created_at").
		First(&record).Error
	if err != nil {
		return entities.User{}, postgres.MapError(err)
	}
	return toUserEntity(record), nil
}

// ByID returns the user together with their role/status in the active
// workspace (resolved from organization_members). The membership role is the
// source of truth for permissions, not the legacy users.role column.
func (r *UserRepository) ByID(ctx context.Context, organizationID, userID string) (entities.User, error) {
	var row struct {
		userModel
		MemberRole   *string `gorm:"column:member_role"`
		MemberStatus *string `gorm:"column:member_status"`
	}
	err := r.db.WithContext(ctx).Table("users AS u").
		Select("u.*, om.role AS member_role, om.status AS member_status").
		Joins("LEFT JOIN organization_members om ON om.user_id = u.id AND om.organization_id = ? AND om.revoked_at IS NULL", organizationID).
		Where("u.id = ? AND u.revoked_at IS NULL", userID).
		First(&row).Error
	if err != nil {
		return entities.User{}, postgres.MapError(err)
	}
	value := toUserEntity(row.userModel)
	if row.MemberRole != nil && *row.MemberRole != "" {
		value.Role = *row.MemberRole
	}
	if row.MemberStatus != nil && *row.MemberStatus != "" {
		value.Status = *row.MemberStatus
	}
	return value, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, principal domain.Principal, firstName, lastName, email string) (entities.User, error) {
	// The user is a global identity; updating it must not be gated by the
	// active workspace's organization_id (which may differ from the user's
	// default workspace). Membership is already enforced upstream.
	result := r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ? AND revoked_at IS NULL", principal.UserID).
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

// AcceptInvite activates an invitation. It reuses an existing global user when
// the invited email already exists, otherwise it provisions a new global user
// with a personal workspace. In both cases an active membership is added to the
// invited workspace. The returned user carries the accepted workspace id/role
// so callers can mint a token that drops the user into the accepted workspace.
func (r *UserRepository) AcceptInvite(ctx context.Context, tokenHash, passwordHash string) (entities.User, error) {
	var invitation invitationModel
	var result entities.User
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
		now := time.Now()
		var existing userModel
		findErr := tx.Where("lower(email) = lower(?) AND revoked_at IS NULL", invitation.Email).
			Order("created_at").First(&existing).Error
		switch {
		case findErr == nil:
			updates := map[string]any{"status": "Aktif", "updated_at": now}
			if existing.PasswordHash == nil || *existing.PasswordHash == "" {
				updates["password_hash"] = passwordHash
			}
			if err := tx.Model(&userModel{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
				return postgres.MapError(err)
			}
			if _, err := addMembershipOnTx(tx, invitation.OrganizationID, existing.ID, invitation.Role, invitation.InvitedBy); err != nil {
				return postgres.MapError(err)
			}
			result = toUserEntity(existing)
		case errors.Is(findErr, gorm.ErrRecordNotFound):
			org := organizationModel{Name: personalWorkspaceName(invitation.Email)}
			if err := tx.Create(&org).Error; err != nil {
				return postgres.MapError(err)
			}
			hash := passwordHash
			first, last := nameFromEmail(invitation.Email)
			record := userModel{
				OrganizationID: org.ID, FirstName: first, LastName: last,
				Email: strings.ToLower(invitation.Email), PasswordHash: &hash,
				Role: domain.RoleOwner, Status: "Aktif",
			}
			if err := tx.Create(&record).Error; err != nil {
				return postgres.MapError(err)
			}
			if err := createOwnerMembership(tx, org.ID, record.ID); err != nil {
				return postgres.MapError(err)
			}
			if err := seedOrganization(tx, org.ID, r.location); err != nil {
				return postgres.MapError(err)
			}
			if _, err := addMembershipOnTx(tx, invitation.OrganizationID, record.ID, invitation.Role, invitation.InvitedBy); err != nil {
				return postgres.MapError(err)
			}
			result = toUserEntity(record)
		default:
			return postgres.MapError(findErr)
		}
		result.OrganizationID = invitation.OrganizationID
		result.Role = invitation.Role
		return postgres.MapError(tx.Model(&invitationModel{}).Where("id = ?", invitation.ID).
			Updates(map[string]any{"accepted_at": now, "user_id": result.ID}).Error)
	})
	if err != nil {
		return entities.User{}, err
	}
	return result, nil
}

func (r *UserRepository) LinkGoogleID(ctx context.Context, organizationID, userID, googleID, avatarURL string) error {
	updates := map[string]any{"google_id": googleID, "updated_at": time.Now()}
	if avatarURL != "" {
		updates["avatar_url"] = avatarURL
	}
	result := r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ? AND revoked_at IS NULL", userID).
		Updates(updates)
	if result.Error != nil {
		return postgres.MapError(result.Error)
	}
	return nil
}

// ListTeam returns the active members of a workspace resolved through
// organization_members (role/status come from the membership, not the user).
func (r *UserRepository) ListTeam(ctx context.Context, organizationID string) ([]entities.User, error) {
	var records []userModel
	if err := r.db.WithContext(ctx).
		Joins("JOIN organization_members om ON om.user_id = users.id AND om.organization_id = ? AND om.revoked_at IS NULL", organizationID).
		Where("users.revoked_at IS NULL").
		Order("om.joined_at NULLS LAST, users.created_at").
		Find(&records).Error; err != nil {
		return nil, err
	}
	// Resolve membership role/status per member.
	members, err := r.teamMemberships(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	items := make([]entities.User, 0, len(records))
	for _, record := range records {
		role, status := record.Role, record.Status
		if m, ok := members[record.ID]; ok {
			role, status = m.Role, m.Status
		}
		items = append(items, toUserEntityWithMembership(record, role, status))
	}
	return items, nil
}

func (r *UserRepository) teamMemberships(ctx context.Context, organizationID string) (map[string]entities.OrganizationMember, error) {
	var rows []organizationMemberModel
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND revoked_at IS NULL", organizationID).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]entities.OrganizationMember, len(rows))
	for _, row := range rows {
		out[row.UserID] = toMemberEntity(row)
	}
	return out, nil
}

// InviteMember records an invitation for an email without creating a user. The
// user is only provisioned (or re-linked) when the invite is accepted.
func (r *UserRepository) InviteMember(ctx context.Context, principal domain.Principal, value entities.User, invitation entities.Invitation) (entities.User, error) {
	invitedBy := principal.UserID
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&invitationModel{
			OrganizationID: principal.OrganizationID,
			Email:          strings.ToLower(value.Email),
			Role:           value.Role,
			TokenHash:      invitation.TokenHash,
			ExpiresAt:      invitation.ExpiresAt,
			InvitedBy:      &invitedBy,
		}).Error; err != nil {
			return postgres.MapError(err)
		}
		return postgres.MapError(tx.Create(&activityModel{
			OrganizationID: principal.OrganizationID, ActorID: principal.UserID,
			ActorName: principal.Name, Action: "mengundang anggota tim", Target: value.Email,
		}).Error)
	})
	if err != nil {
		return entities.User{}, err
	}
	first, last := splitInviteName(value.FirstName, value.LastName, value.Email)
	return entities.User{
		FirstName: first, LastName: last, Email: strings.ToLower(value.Email),
		Role: value.Role, Status: "Menunggu",
	}, nil
}

// RevokeMember revokes the user's membership in the workspace. It does not
// delete the global user identity, so the member keeps access to any other
// workspaces they own or belong to.
func (r *UserRepository) RevokeMember(ctx context.Context, organizationID, userID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&organizationMemberModel{}).
		Where("organization_id = ? AND user_id = ? AND revoked_at IS NULL", organizationID, userID).
		Updates(map[string]any{"status": "Dicabut", "revoked_at": now, "updated_at": now})
	if result.Error != nil {
		return postgres.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// --- workspace-aware membership methods ---

// ListWorkspaces returns every active workspace a user can access, with the
// user's default workspace (users.organization_id) listed first.
func (r *UserRepository) ListWorkspaces(ctx context.Context, userID string) ([]entities.WorkspaceSummary, error) {
	var rows []struct {
		ID        string `gorm:"column:id"`
		Name      string `gorm:"column:name"`
		Role      string `gorm:"column:role"`
		IsDefault bool   `gorm:"column:is_default"`
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT o.id AS id, o.name AS name, om.role AS role,
		       (u.organization_id = o.id) AS is_default
		FROM organization_members om
		JOIN organizations o ON o.id = om.organization_id
		JOIN users u ON u.id = om.user_id
		WHERE om.user_id = ? AND om.status = 'Aktif' AND om.revoked_at IS NULL
		ORDER BY is_default DESC, om.joined_at NULLS LAST, om.created_at`, userID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]entities.WorkspaceSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, entities.WorkspaceSummary{
			ID: row.ID, Name: row.Name, Role: row.Role, IsDefault: row.IsDefault,
		})
	}
	return items, nil
}

// Membership returns the active membership of a user in a workspace.
func (r *UserRepository) Membership(ctx context.Context, userID, organizationID string) (entities.OrganizationMember, error) {
	var record organizationMemberModel
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND organization_id = ? AND status = 'Aktif' AND revoked_at IS NULL", userID, organizationID).
		First(&record).Error
	if err != nil {
		return entities.OrganizationMember{}, postgres.MapError(err)
	}
	return toMemberEntity(record), nil
}

// CreatePersonalWorkspace provisions a new workspace owned by an existing user
// and seeds its pipeline data. The user's default workspace pointer is updated.
func (r *UserRepository) CreatePersonalWorkspace(ctx context.Context, orgName string, user entities.User) (entities.OrganizationMember, error) {
	var member organizationMemberModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		org := organizationModel{Name: orgName}
		if err := tx.Create(&org).Error; err != nil {
			return postgres.MapError(err)
		}
		if err := createOwnerMembership(tx, org.ID, user.ID); err != nil {
			return postgres.MapError(err)
		}
		if err := seedOrganization(tx, org.ID, r.location); err != nil {
			return postgres.MapError(err)
		}
		if err := tx.Model(&userModel{}).Where("id = ?", user.ID).
			Updates(map[string]any{"organization_id": org.ID, "role": domain.RoleOwner, "status": "Aktif", "updated_at": time.Now()}).Error; err != nil {
			return postgres.MapError(err)
		}
		return tx.Where("organization_id = ? AND user_id = ?", org.ID, user.ID).First(&member).Error
	})
	if err != nil {
		return entities.OrganizationMember{}, postgres.MapError(err)
	}
	return toMemberEntity(member), nil
}

// AddMembership inserts (or reactivates) an active membership for a user in a
// workspace.
func (r *UserRepository) AddMembership(ctx context.Context, organizationID, userID, role string, invitedBy *string) (entities.OrganizationMember, error) {
	var member organizationMemberModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m, err := addMembershipOnTx(tx, organizationID, userID, role, invitedBy)
		if err != nil {
			return err
		}
		member = m
		return nil
	})
	if err != nil {
		return entities.OrganizationMember{}, postgres.MapError(err)
	}
	return toMemberEntity(member), nil
}

// LookupInvitation loads an invitation by its token hash.
func (r *UserRepository) LookupInvitation(ctx context.Context, tokenHash string) (entities.Invitation, error) {
	var record invitationModel
	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&record).Error
	if err != nil {
		return entities.Invitation{}, postgres.MapError(err)
	}
	return toInvitationEntity(record), nil
}

// MarkInvitationAccepted records the acceptance timestamp for an invitation.
func (r *UserRepository) MarkInvitationAccepted(ctx context.Context, invitationID, accepterUserID string) error {
	result := r.db.WithContext(ctx).Model(&invitationModel{}).
		Where("id = ? AND accepted_at IS NULL", invitationID).
		Updates(map[string]any{"accepted_at": time.Now(), "user_id": accepterUserID})
	if result.Error != nil {
		return postgres.MapError(result.Error)
	}
	return nil
}

// ActiveMembershipExists reports whether the email already has an active
// membership in the workspace.
func (r *UserRepository) ActiveMembershipExists(ctx context.Context, organizationID, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("organization_members AS om").
		Joins("JOIN users u ON u.id = om.user_id").
		Where("om.organization_id = ? AND lower(u.email) = lower(?) AND om.status = 'Aktif' AND om.revoked_at IS NULL", organizationID, email).
		Count(&count).Error
	return count > 0, err
}

// --- helpers ---

// createOwnerMembership inserts an active Owner membership row.
func createOwnerMembership(tx *gorm.DB, organizationID, userID string) error {
	now := time.Now()
	return tx.Create(&organizationMemberModel{
		OrganizationID: organizationID, UserID: userID, Role: domain.RoleOwner,
		Status: "Aktif", JoinedAt: &now,
	}).Error
}

// addMembershipOnTx upserts an active membership within an existing
// transaction. If a (possibly revoked) membership exists it is reactivated with
// the new role; otherwise a new row is inserted.
func addMembershipOnTx(tx *gorm.DB, organizationID, userID, role string, invitedBy *string) (organizationMemberModel, error) {
	var member organizationMemberModel
	err := tx.Where("organization_id = ? AND user_id = ?", organizationID, userID).First(&member).Error
	now := time.Now()
	switch {
	case err == nil:
		updates := map[string]any{
			"role": role, "status": "Aktif", "revoked_at": nil,
			"joined_at": now, "updated_at": now,
		}
		if invitedBy != nil {
			updates["invited_by"] = *invitedBy
		}
		if err := tx.Model(&member).Updates(updates).Error; err != nil {
			return organizationMemberModel{}, err
		}
		member.Role = role
		member.Status = "Aktif"
		member.RevokedAt = nil
		return member, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		member = organizationMemberModel{
			OrganizationID: organizationID, UserID: userID, Role: role,
			Status: "Aktif", InvitedBy: invitedBy, JoinedAt: &now,
		}
		if err := tx.Create(&member).Error; err != nil {
			return organizationMemberModel{}, err
		}
		return member, nil
	default:
		return organizationMemberModel{}, err
	}
}

// seedOrganization creates the default pipeline stages and performance goals
// for a freshly created workspace.
func seedOrganization(tx *gorm.DB, organizationID string, location *time.Location) error {
	stages := []stageModel{
		{OrganizationID: organizationID, Key: "lead", Name: "Lead Masuk", Color: "bg-primary-container", Position: 1, IsSystem: true},
		{OrganizationID: organizationID, Key: "contacted", Name: "Dihubungi", Color: "bg-secondary-container", Position: 2, IsSystem: true},
		{OrganizationID: organizationID, Key: "meeting", Name: "Meeting", Color: "bg-tertiary-container", Position: 3, IsSystem: true},
		{OrganizationID: organizationID, Key: "negotiation", Name: "Negosiasi", Color: "bg-primary-fixed", Position: 4, IsSystem: true},
		{OrganizationID: organizationID, Key: "won", Name: "Deal Won", Color: "bg-surface-tint", Position: 5, IsSystem: true},
		{OrganizationID: organizationID, Key: "lost", Name: "Deal Lost", Color: "bg-error-container", Position: 6, IsSystem: true},
	}
	if err := tx.Create(&stages).Error; err != nil {
		return err
	}
	now := time.Now().In(location)
	goals := make([]goalModel, 0, 3)
	for offset, target := range []int64{1_000_000_000, 900_000_000, 900_000_000} {
		goals = append(goals, goalModel{
			OrganizationID: organizationID,
			Month:          time.Date(now.Year(), now.Month()-time.Month(offset), 1, 0, 0, 0, 0, location),
			Goal:           target,
		})
	}
	return tx.Create(&goals).Error
}

func personalWorkspaceName(email string) string {
	if at := strings.IndexByte(email, '@'); at > 0 {
		return "CRM " + email[:at]
	}
	return "CRM " + email
}

func nameFromEmail(email string) (string, string) {
	if at := strings.IndexByte(email, '@'); at > 0 {
		return email[:at], ""
	}
	return email, ""
}

func splitInviteName(first, last, email string) (string, string) {
	if strings.TrimSpace(first) != "" || strings.TrimSpace(last) != "" {
		return first, last
	}
	return nameFromEmail(email)
}

func toMemberEntity(record organizationMemberModel) entities.OrganizationMember {
	return entities.OrganizationMember{
		ID: record.ID, OrganizationID: record.OrganizationID, UserID: record.UserID,
		Role: record.Role, Status: record.Status, InvitedBy: record.InvitedBy,
		JoinedAt: record.JoinedAt, RevokedAt: record.RevokedAt,
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}
}

func toInvitationEntity(record invitationModel) entities.Invitation {
	return entities.Invitation{
		ID: record.ID, OrganizationID: record.OrganizationID, UserID: record.UserID,
		Email: record.Email, Role: record.Role, TokenHash: record.TokenHash,
		ExpiresAt: record.ExpiresAt, AcceptedAt: record.AcceptedAt,
		InvitedBy: record.InvitedBy, CreatedAt: record.CreatedAt,
	}
}

func toUserEntity(record userModel) entities.User {
	passwordHash := ""
	if record.PasswordHash != nil {
		passwordHash = *record.PasswordHash
	}
	value := entities.User{
		ID: record.ID, OrganizationID: record.OrganizationID,
		FirstName: record.FirstName, LastName: record.LastName, Email: record.Email,
		PasswordHash: passwordHash, GoogleID: record.GoogleID,
		Role: record.Role, Status: record.Status,
		AvatarURL: record.AvatarURL, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}
	value.Name = strings.TrimSpace(value.FirstName + " " + value.LastName)
	value.Initials = postgres.Initials(value.Name)
	return value
}

func toUserEntityWithMembership(record userModel, role, status string) entities.User {
	value := toUserEntity(record)
	value.Role = role
	value.Status = status
	return value
}
