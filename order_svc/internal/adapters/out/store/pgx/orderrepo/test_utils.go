package orderrepo

import (
	"context"
	"path/filepath"
	"time"

	"github.com/Anacardo89/order_svc_hex/order_svc/config"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/db"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/testutils"
)

func InitDB(ctx context.Context, dsn string) (core.OrderRepo, error) {
	dbCfg := config.DB{
		DSN:             dsn,
		MaxConns:        5,
		MinConns:        1,
		MaxConnLifetime: 2 * time.Minute,
		MaxConnIdleTime: 2 * time.Minute,
	}
	dbConn, err := db.Connect(dbCfg)
	if err != nil {
		return nil, err
	}
	repo := NewRepo(dbConn)
	return repo, nil
}

func BuildTestDBEnv(ctx context.Context) (core.OrderRepo, string, func(), string, error) {
	dsn, closeDB, err := testutils.StartPostgresContainer(ctx)
	if err != nil {
		return nil, "", nil, "", err
	}
	repo, err := InitDB(ctx, dsn)
	if err != nil {
		closeDB()
		return nil, "", nil, "", err
	}
	migratePath := filepath.Join("db", "migrations")
	migratePath, err = testutils.BuildPath(migratePath)
	if err != nil {
		closeDB()
		return nil, "", nil, "", err
	}
	if err := db.Migrate(dsn, migratePath, db.MigrateUp); err != nil {
		closeDB()
		return nil, "", nil, "", err
	}
	seedPath := filepath.Join("db", "seeds")
	seedPath, err = testutils.BuildPath(seedPath)
	if err != nil {
		closeDB()
		return nil, "", nil, "", err
	}
	return repo, dsn, closeDB, seedPath, nil
}
