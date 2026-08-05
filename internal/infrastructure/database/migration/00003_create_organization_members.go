package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"gorm.io/gorm"
)

// CreateOrganizationMembers introduces the many-to-many membership table that
// allows a single user to belong to multiple workspaces. Existing users are
// backfilled as members of their current organization. The legacy columns on
// users (organization_id, role, status) are intentionally kept for backwards
// compatibility; membership becomes the new source of truth.
var CreateOrganizationMembers = &gormigrate.Migration{
	ID: "00003_create_organization_members",
	Migrate: func(tx *gorm.DB) error {
		if !tx.Migrator().HasTable(&entities.OrganizationMember{}) {
			if err := tx.Migrator().CreateTable(&entities.OrganizationMember{}); err != nil {
				return err
			}
		}

		// Unique constraint so a user can only have one membership per workspace.
		if !tx.Migrator().HasIndex(&entities.OrganizationMember{}, "idx_organization_members_org_user") {
			if err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_organization_members_org_user
				ON organization_members (organization_id, user_id)`).Error; err != nil {
				return err
			}
		}

		// Index for listing a user's active memberships.
		if !tx.Migrator().HasIndex(&entities.OrganizationMember{}, "idx_organization_members_user_status") {
			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_organization_members_user_status
				ON organization_members (user_id, status)`).Error; err != nil {
				return err
			}
		}

		// Index for listing a workspace's members.
		if !tx.Migrator().HasIndex(&entities.OrganizationMember{}, "idx_organization_members_org_status") {
			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_organization_members_org_status
				ON organization_members (organization_id, status)`).Error; err != nil {
				return err
			}
		}

		// user_invitations.user_id becomes nullable: invitations no longer create
		// a user upfront. invited_by records who sent the invite.
		if tx.Migrator().HasTable("user_invitations") {
			if tx.Migrator().HasColumn("user_invitations", "user_id") {
				if err := tx.Exec(`ALTER TABLE user_invitations ALTER COLUMN user_id DROP NOT NULL`).Error; err != nil {
					return err
				}
			}
			if !tx.Migrator().HasColumn("user_invitations", "invited_by") {
				if err := tx.Exec(`ALTER TABLE user_invitations ADD COLUMN IF NOT EXISTS invited_by uuid`).Error; err != nil {
					return err
				}
			}
		}

		// Backfill: every user with an organization becomes a member. Historical
		// Admin users are promoted to Owner of their own workspace.
		return tx.Exec(`
			INSERT INTO organization_members (organization_id, user_id, role, status, joined_at, created_at, updated_at)
			SELECT u.organization_id, u.id,
				CASE WHEN u.role = 'Admin' THEN 'Owner' ELSE u.role END,
				u.status,
				u.created_at, u.created_at, u.created_at
			FROM users u
			WHERE u.organization_id IS NOT NULL
			ON CONFLICT DO NOTHING`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable(&entities.OrganizationMember{})
	},
}
