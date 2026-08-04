package main

import (
	"log"
	"os"

	"github.com/prayogopangestu/crm-system/backend/configs"
	migrations "github.com/prayogopangestu/crm-system/backend/internal/infrastructure/database/migration"
	infraPostgres "github.com/prayogopangestu/crm-system/backend/internal/infrastructure/database/postgres"
	"github.com/prayogopangestu/crm-system/backend/pkg/envfile"

	"github.com/go-gormigrate/gormigrate/v2"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	envfile.Load(".env")

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	cfg, err := configs.Load(configPath)
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	sqlDB, err := infraPostgres.NewConnection(infraPostgres.Config{
		URL:          cfg.Database.URL,
		MaxOpenConns: 5,
		MaxIdleConns: 5,
	})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer sqlDB.Close()

	db, err := gorm.Open(gormpostgres.New(gormpostgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to open gorm connection:", err)
	}

	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		migrations.CreateInitialSchema,
		migrations.AddGoogleID,
		migrations.CreateOrganizationMembers,
	})

	if err := m.Migrate(); err != nil {
		log.Fatalf("could not migrate: %v", err)
	}

	log.Println("migration run successfully")
}
