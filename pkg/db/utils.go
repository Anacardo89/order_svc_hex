package db

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

type MigrateDirection string

const (
	MigrateUp   MigrateDirection = "up"
	MigrateDown MigrateDirection = "down"
)

func MigrateDB(dsn, migrationsPath string, direction MigrateDirection) error {
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migration directory not found at: %s", migrationsPath)
	}
	m, err := migrate.New(fmt.Sprintf("file://%s", migrationsPath), dsn)
	if err != nil {
		return fmt.Errorf("error migrating: %s", err.Error())
	}
	switch direction {
	case MigrateUp:
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	case MigrateDown:
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	default:
		return fmt.Errorf("unknown migration direction: %s", direction)
	}
	return nil
}

func SeedDB(ctx context.Context, gormDB *gorm.DB, seedPath string) error {
	seed, err := os.ReadFile(seedPath)
	if err != nil {
		return fmt.Errorf("failed to read seed file: %w", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	_, err = sqlDB.Exec(string(seed))
	if err != nil {
		return fmt.Errorf("failed to execute seed: %w", err)
	}
	return nil
}
