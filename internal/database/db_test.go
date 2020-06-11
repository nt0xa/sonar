package database_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/database"
)

var (
	db       *database.DB
	fixtures *testfixtures.Context
)

func TestMain(m *testing.M) {
	if err := setupGlobals(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	ret := m.Run()

	if err := teardownGlobals(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(ret)
}

func setupGlobals() error {
	cfg := &database.Config{
		DSN:        os.Getenv("SONAR_DB_DSN"),
		Migrations: "migrations",
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	var err error

	db, err = database.New(cfg)
	if err != nil {
		return fmt.Errorf("fail to init db: %w", err)
	}

	if err := db.Migrate(); err != nil {
		return fmt.Errorf("fail to apply migrations: %w", err)
	}

	fixtures, err = testfixtures.NewFolder(db.DB.DB,
		&testfixtures.PostgreSQL{}, "fixtures")
	if err != nil {
		return fmt.Errorf("fail to load fixtures: %w", err)
	}

	return nil
}

func teardownGlobals() error {
	if db != nil {
		if err := db.Close(); err != nil {
			return fmt.Errorf("fail to close connection: %w", err)
		}
	}
	return nil
}

func setup(t *testing.T) {
	err := fixtures.Load()
	require.NoError(t, err)
}

func teardown(t *testing.T) {}
