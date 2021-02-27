package dnsx_test

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-testfixtures/testfixtures"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/handlers/dnsx"
)

var (
	srv      *dnsx.Server
	db       *database.DB
	tf       *testfixtures.Context
	notifier *NotifierMock
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

	db, err = database.New(&database.Config{
		DSN:        dsn,
		Migrations: "../../database/migrations",
	})
	if err != nil {
		return fmt.Errorf("fail to init db: %w", err)
	}

	if err := db.Migrate(); err != nil {
		return fmt.Errorf("fail to apply migrations: %w", err)
	}

	tf, err = testfixtures.NewFolder(
		db.DB.DB,
		&testfixtures.PostgreSQL{},
		"../../database/fixtures",
	)
	if err != nil {
		return fmt.Errorf("fail to load fixtures: %w", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	notifier = &NotifierMock{}
	domain := "sonar.local"

	defaultRecords, err := dnsx.DefaultRecords(domain, net.ParseIP("127.0.0.1"))
	if err != nil {
		return fmt.Errorf("fail to init default dns records: %w", err)
	}

	srv = dnsx.New(":1053", domain,
		[]dnsx.Finder{
			dnsx.NewDatabaseFinder(db, domain),
			defaultRecords,
		},
		dnsx.NotifyRequestFunc(notifier.Notify),
		dnsx.NotifyStartedFunc(func() {
			wg.Done()
		}),
	)
	if err != nil {
		return err
	}

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

type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) Notify(remoteAddr net.Addr, data []byte, meta map[string]interface{}) {
	m.Called(remoteAddr, data, meta)
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}

func setup(t *testing.T) {
	err := tf.Load()
	require.NoError(t, err)
}

func teardown(t *testing.T) {}
