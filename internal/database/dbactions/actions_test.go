package dbactions_test

import (
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/testutils"
)

var (
	db   *database.DB
	tf   *testfixtures.Context
	acts actions.Actions

	g = testutils.Globals(
		testutils.DB(&database.Config{
			DSN:        os.Getenv("SONAR_DB_DSN"),
			Migrations: "../migrations",
		}, &db),
		testutils.Fixtures(&db, "../fixtures", &tf),
		testutils.ActionsDB(&db, &acts),
	)
)

func TestMain(m *testing.M) {
	testutils.TestMain(m, g)
}

func setup(t *testing.T) {
	err := tf.Load()
	require.NoError(t, err)
}

func teardown(t *testing.T) {}
