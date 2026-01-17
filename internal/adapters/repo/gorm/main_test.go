package gormrepo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var (
	dsn      string
	seedPath string
	repo     *OrderRepo
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	var (
		closeDB func()
		err     error
	)
	repo, dsn, closeDB, seedPath, err = BuildTestDBEnv(ctx)
	if err != nil {
		fmt.Println("Failed to start test environment:", err)
		os.Exit(1)
	}
	defer closeDB()
	seedPath = filepath.Join(seedPath, "orders_test.sql")
	code := m.Run()
	os.Exit(code)
}
