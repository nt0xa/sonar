package api_test

import (
	"errors"
	"flag"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/go-testfixtures/testfixtures"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/modules/api"
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
	cfg *api.Config
	db  *database.DB
	srv *httptest.Server
	tf  *testfixtures.Context
)

const (
	AdminToken = "94008eb13da98b94b5933cd1bd15a359"
	User1Token = "50c862e41d059eeca13adc7b276b46b7"
	User2Token = "7001f2d819d3d5fb0b1fd75dd38eb34e"
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
		Migrations: "../../database/migrations",
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
		"../../database/fixtures",
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
	actions := actions.New(db, log)

	// API controller
	api, err := api.New(cfg, db, log, nil, actions)
	if err != nil {
		return err
	}

	// Create httptest server
	srv = httptest.NewServer(api.Router())

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

//
// httpexpect helpers
//

func heDefault(t *testing.T) *httpexpect.Expect {
	printers := make([]httpexpect.Printer, 0)

	if verbose {
		printers = append(printers, httpexpect.NewCurlPrinter(t))
		printers = append(printers, httpexpect.NewDebugPrinter(t, true))
	}

	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  srv.URL,
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: printers,
	})
}

func heAuth(he *httpexpect.Expect, token string) *httpexpect.Expect {
	return he.Builder(func(r *httpexpect.Request) {
		r.WithHeader("Authorization", fmt.Sprintf("Bearer %s", token))
	})
}
