package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
	"gorm.io/gorm"
)

var CreateInitialSchema = &gormigrate.Migration{
	ID: "00001_create_initial_schema",
	Migrate: func(tx *gorm.DB) error {
		entitiesToCreate := []interface{}{
			&entities.Organization{},
			&entities.User{},
			&entities.Invitation{},
			&entities.Contact{},
			&entities.Stage{},
			&entities.Deal{},
			&entities.Task{},
			&entities.Activity{},
			&entities.PerformanceGoal{},
			&entities.Notification{},
			&entities.Telegram{},
			&entities.OutboxEvent{},
		}
		for _, model := range entitiesToCreate {
			if !tx.Migrator().HasTable(model) {
				if err := tx.Migrator().CreateTable(model); err != nil {
					return err
				}
			}
		}
		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable(
			&entities.OutboxEvent{},
			&entities.Telegram{},
			&entities.Notification{},
			&entities.PerformanceGoal{},
			&entities.Activity{},
			&entities.Task{},
			&entities.Deal{},
			&entities.Stage{},
			&entities.Contact{},
			&entities.Invitation{},
			&entities.User{},
			&entities.Organization{},
		)
	},
}
