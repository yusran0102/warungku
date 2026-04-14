package database

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect opens the PostgreSQL connection used by GORM for all queries.
func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	fmt.Println("✅ Database connected")
}

// RunMigrations applies all pending UP migrations from db/migrations/.
// It is safe to call on every app start — golang-migrate tracks applied
// migrations in a "schema_migrations" table it manages automatically.
func RunMigrations() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	m, err := migrate.New("file://db/migrations", dsn)
	if err != nil {
		log.Fatalf("migrate: failed to create instance: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("✅ Migrations: nothing new to apply")
			return
		}
		log.Fatalf("migrate: failed to apply migrations: %v", err)
	}

	fmt.Println("✅ Migrations applied successfully")
}

// MigrateDown rolls back the last N steps. Use from CLI only.
func MigrateDown(steps int) {
	dsn := os.Getenv("DATABASE_URL")
	m, err := migrate.New("file://db/migrations", dsn)
	if err != nil {
		log.Fatalf("migrate: %v", err)
	}
	defer m.Close()

	if err := m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migrate down: %v", err)
	}
	fmt.Printf("✅ Rolled back %d migration(s)\n", steps)
}

// MigrateVersion prints the current migration version.
func MigrateVersion() {
	dsn := os.Getenv("DATABASE_URL")
	m, err := migrate.New("file://db/migrations", dsn)
	if err != nil {
		log.Fatalf("migrate: %v", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("No migrations applied yet")
			return
		}
		log.Fatalf("migrate version: %v", err)
	}
	fmt.Printf("Current version: %d (dirty: %v)\n", version, dirty)
}
