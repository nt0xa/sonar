package actionsdb_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/actionsdb"
	"github.com/nt0xa/sonar/internal/database"
)

var (
	tf   *testfixtures.Loader
	db   *database.DB
	acts actions.Actions
)

func TestMain(m *testing.M) {
	var (
		dsn string
		err error
	)

	if dsn = os.Getenv("SONAR_DB_DSN"); dsn == "" {
		fmt.Fprintln(os.Stderr, "empty SONAR_DB_DSN")
		os.Exit(1)
	}

	log := logrus.New()

	db, err = database.New(&database.Config{DSN: dsn}, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to init database: %v\n", err)
		os.Exit(1)
	}

	if err := db.Migrate(); err != nil {
		fmt.Fprintf(os.Stderr, "fail to apply database migrations: %v\n", err)
		os.Exit(1)
	}

	acts = actionsdb.New(db, log, "sonar.test")

	tf, err = testfixtures.New(
		testfixtures.Database(db.DB.DB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("../database/fixtures"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to load fixtures: %v", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func setup(t *testing.T) {
	err := tf.Load()
	require.NoError(t, err)
}

func teardown(t *testing.T) {}
