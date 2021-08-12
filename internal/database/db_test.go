package database_test

import (
	"testing"

	"github.com/go-testfixtures/testfixtures"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/testutils"
)

var (
	db  *database.DB
	tf  *testfixtures.Context
	log = logrus.New()

	g = testutils.Globals(
		testutils.DB(&db, log),
		testutils.Fixtures(&db, &tf),
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
