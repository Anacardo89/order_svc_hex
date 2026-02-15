package testutils

import (
	"context"
	"fmt"
	"os"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
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

func ConnectTestDB(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	return pool, nil
}

func SeedTestDB(ctx context.Context, pool *pgxpool.Pool, seedPath string) error {
	seedSQL, err := os.ReadFile(seedPath)
	if err != nil {
		return fmt.Errorf("failed to read seed file: %w", err)
	}
	_, err = pool.Exec(ctx, string(seedSQL))
	if err != nil {
		return fmt.Errorf("failed to execute seed SQL: %w", err)
	}
	return nil
}
