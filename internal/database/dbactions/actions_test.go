package dbactions_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/database/dbactions"
)

func TestMain(m *testing.M) {
	if err := Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	ret := m.Run()

	if err := Teardown(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(ret)
}

//
// Setup & Teardown globals
//

var (
	db   *database.DB
	tf   *testfixtures.Context
	acts actions.Actions
)

func Setup() error {
	var err error

	dsn, ok := os.LookupEnv("SONAR_DB_DSN")
	if !ok {
		return errors.New("empty SONAR_DB_DSN")
	}

	// DB
	db, err = database.New(&database.Config{
		DSN:        dsn,
		Migrations: "../migrations",
	})
	if err != nil {
		return fmt.Errorf("fail to init db: %w", err)
	}

	// Migrations
	if err := db.Migrate(); err != nil {
		return fmt.Errorf("fail to apply migrations: %w", err)
	}

	// Load DB fixtures
	tf, err = testfixtures.NewFolder(
		db.DB.DB,
		&testfixtures.PostgreSQL{},
		"../fixtures",
	)
	if err != nil {
		return fmt.Errorf("fail to load fixtures: %w", err)
	}

	// Logger
	log := logrus.New()

	// Actions
	acts = dbactions.New(db, log, "sonar.local")

	return nil
}

func Teardown() error {
	// Close database connection
	if db != nil {
		if err := db.Close(); err != nil {
			return fmt.Errorf("model: fail to close: %w", err)
		}
	}
	return nil
}

//
// setup & teardown
//

func setup(t *testing.T) {
	err := tf.Load()
	require.NoError(t, err)
}

func teardown(t *testing.T) {}
