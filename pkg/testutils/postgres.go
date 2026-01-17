package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	testcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func StartPostgresContainer(ctx context.Context) (string, func(), error) {
	pgContainer, err := testcontainer.Run(ctx, "postgres:16-alpine",
		testcontainer.WithDatabase("testdb"),
		testcontainer.WithUsername("test"),
		testcontainer.WithPassword("secret"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return "", nil, err
	}
	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		pgContainer.Terminate(ctx)
		return "", nil, err
	}
	close := func() { _ = pgContainer.Terminate(ctx) }
	return dsn, close, nil
}

func ConnectTestDB(dsn string) (*sql.DB, error) {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	// Ping
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping test DB: %w", err)
	}
	return sqlDB, nil
}
