package apiclient_test

import (
	"errors"
	"flag"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/database/dbactions"
	"github.com/bi-zone/sonar/internal/modules/api"
	"github.com/bi-zone/sonar/internal/modules/api/apiclient"
)

// Flags
var (
	logs    bool
	verbose bool
)

var _ = func() bool {
	testing.Init()
	return true
}()

func init() {
	flag.BoolVar(&logs, "test.logs", false, "Enables logger output.")
	flag.BoolVar(&verbose, "test.verbose", false, "Enables verbose HTTP printing.")
	flag.Parse()
}

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
	cfg    *api.Config
	db     *database.DB
	srv    *httptest.Server
	tf     *testfixtures.Context
	client *apiclient.Client
)

const (
	AdminToken = "94008eb13da98b94b5933cd1bd15a359"
	UserToken  = "50c862e41d059eeca13adc7b276b46b7"
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
		Migrations: "../../../database/migrations",
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
		"../../../database/fixtures",
	)
	if err != nil {
		return fmt.Errorf("fail to load fixtures: %w", err)
	}

	// Config
	cfg = &api.Config{
		Admin: AdminToken,
	}

	// Logger
	log := logrus.New()

	// Actions
	actions := dbactions.New(db, log, "sonar.local")

	// API controller
	api, err := api.New(cfg, db, log, nil, actions)
	if err != nil {
		return err
	}

	// Create httptest server
	srv = httptest.NewServer(api.Router())

	// Create api client
	client = apiclient.New(srv.URL, UserToken, false)

	return nil
}

func Teardown() error {
	// Close database connection
	if db != nil {
		if err := db.Close(); err != nil {
			return fmt.Errorf("model: fail to close: %w", err)
		}
	}

	// Stop httptest server
	if srv != nil {
		srv.Close()
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
