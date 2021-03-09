package actionsdb_test

import (
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/testutils"
)

var (
	db   *database.DB
	tf   *testfixtures.Context
	acts actions.Actions

	log = logrus.New()

	g = testutils.Globals(
		testutils.DB(&database.Config{
			DSN:        os.Getenv("SONAR_DB_DSN"),
			Migrations: "../database/migrations",
		}, &db),
		testutils.Fixtures(&db, "../database/fixtures", &tf),
		testutils.ActionsDB(&db, log, &acts),
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
