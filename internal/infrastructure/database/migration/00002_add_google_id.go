package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"gorm.io/gorm"
)

var AddGoogleID = &gormigrate.Migration{
	ID: "00002_add_google_id",
	Migrate: func(tx *gorm.DB) error {
		if !tx.Migrator().HasColumn(&entities.User{}, "google_id") {
			if err := tx.Migrator().AddColumn(&entities.User{}, "GoogleID"); err != nil {
				return err
			}
		}
		if !tx.Migrator().HasIndex(&entities.User{}, "google_id") {
			if err := tx.Migrator().CreateIndex(&entities.User{}, "GoogleID"); err != nil {
				return err
			}
		}
		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		_ = tx.Migrator().DropIndex(&entities.User{}, "google_id")
		return tx.Migrator().DropColumn(&entities.User{}, "GoogleID")
	},
}
