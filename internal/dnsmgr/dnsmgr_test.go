package dnsmgr_test

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/dnsmgr"
	"github.com/bi-zone/sonar/pkg/server/dns"
	"github.com/go-testfixtures/testfixtures"
	"github.com/stretchr/testify/require"
)

var (
	mgr *dnsmgr.DNSMgr
	db  *database.DB
	tf  *testfixtures.Context
)

func TestMain(m *testing.M) {
	if err := setupGlobals(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	ret := m.Run()

	os.Exit(ret)
}

func setupGlobals() error {
	var err error

	dsn, ok := os.LookupEnv("SONAR_DB_DSN")
	if !ok {
		return errors.New("empty SONAR_DB_DSN")
	}

	// DB
	db, err = database.New(&database.Config{
		DSN:        dsn,
		Migrations: "../database/migrations",
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
		"../database/fixtures",
	)
	if err != nil {
		return fmt.Errorf("fail to load fixtures: %w", err)
	}

	mgr, err = dnsmgr.New("sonar.local", net.ParseIP("127.0.0.1"), "[a-f0-9]{8}", db)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	srv := dns.New(":1053", mgr.HandleFunc,
		dns.NotifyStartedFunc(func() {
			wg.Done()
		}))

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(fmt.Errorf("fail to start server: %w", err))
		}
	}()

	if waitTimeout(&wg, 30*time.Second) {
		return errors.New("timeout waiting for server to start")
	}

	return nil
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func setup(t *testing.T) {
	err := tf.Load()
	require.NoError(t, err)
}

func teardown(t *testing.T) {}
