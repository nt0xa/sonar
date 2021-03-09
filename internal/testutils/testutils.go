package testutils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-testfixtures/testfixtures"
	"github.com/stretchr/testify/mock"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/actionsdb"
	"github.com/bi-zone/sonar/internal/cmd/server"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/modules/api"
	"github.com/bi-zone/sonar/internal/modules/api/apiclient"
	"github.com/bi-zone/sonar/internal/utils/logger"
	"github.com/bi-zone/sonar/pkg/dnsx"
	"github.com/bi-zone/sonar/pkg/httpx"
)

type Global interface {
	Setup() error
	Teardown() error
}

type global struct {
	setup    func() error
	teardown func() error
}

func (g *global) Setup() error {
	if g.setup != nil {
		return g.setup()
	}

	return nil
}

func (g *global) Teardown() error {
	if g.teardown != nil {
		return g.teardown()
	}

	return nil
}

type globals []Global

func Globals(g ...Global) Global {
	return globals(g)
}

func (g globals) Setup() error {
	for _, it := range g {
		if err := it.Setup(); err != nil {
			return err
		}
	}

	return nil
}

func (g globals) Teardown() error {
	for _, it := range g {
		if err := it.Teardown(); err != nil {
			return err
		}
	}

	return nil
}

func TestMain(m *testing.M, g Global) {
	if err := g.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	ret := m.Run()

	if err := g.Teardown(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(ret)
}

type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) Notify(remoteAddr net.Addr, data []byte, meta map[string]interface{}) {
	m.Called(remoteAddr, data, meta)
}

var (
	TestDomain = "sonar.local"
	TestIP     = net.ParseIP("127.0.0.1")
)

func DB(cfg *database.Config, out **database.DB) Global {
	return &global{
		setup: func() error {
			if err := cfg.Validate(); err != nil {
				return err
			}

			db, err := database.New(cfg)
			if err != nil {
				return fmt.Errorf("fail to init database: %w", err)
			}

			if err := db.Migrate(); err != nil {
				return fmt.Errorf("fail to apply database migrations: %w", err)
			}

			*out = db

			return nil
		},
		teardown: func() error {
			if out == nil {
				return nil
			}

			return (*out).DB.Close()
		},
	}
}

func Fixtures(db **database.DB, path string, out **testfixtures.Context) Global {
	return &global{
		setup: func() error {
			fixtures, err := testfixtures.NewFolder((*db).DB.DB, &testfixtures.PostgreSQL{}, path)
			if err != nil {
				return fmt.Errorf("fail to load fixtures: %w", err)
			}

			*out = fixtures

			return nil
		},
	}
}

func ActionsDB(db **database.DB, log logger.StdLogger, out *actions.Actions) Global {
	return &global{
		setup: func() error {
			*out = actionsdb.New(*db, log, TestDomain)
			return nil
		},
	}
}

func APIServer(cfg *api.Config, db **database.DB, log logger.StdLogger, acts *actions.Actions, out **httptest.Server) Global {
	return &global{
		setup: func() error {
			api, err := api.New(cfg, *db, log, nil, *acts)
			if err != nil {
				return err
			}

			*out = httptest.NewServer(api.Router())
			return nil
		},
	}
}

func APIClient(srv **httptest.Server, token string, out **apiclient.Client) Global {
	return &global{
		setup: func() error {
			*out = apiclient.New((*srv).URL, token, false)
			return nil
		},
	}
}

func DNSX(db **database.DB, notify func(net.Addr, []byte, map[string]interface{}), h *dnsx.HandlerProvider, srv *dnsx.Server) Global {
	return &global{
		setup: func() error {
			wait := sync.Mutex{}

			*h = server.DNSHandler(*db, TestDomain, net.ParseIP("127.0.0.1"), notify)
			*srv = dnsx.New(":1053", *h, dnsx.NotifyStartedFunc(wait.Unlock))

			go func() {
				if err := (*srv).ListenAndServe(); err != nil {
					log.Fatal(fmt.Errorf("fail to start server: %w", err))
				}
			}()

			if waitTimeout(&wait, 30*time.Second) {
				return errors.New("timeout waiting for server to start")
			}

			return nil
		},
	}
}

func TLSConfig(cert, key string, out **tls.Config) Global {
	return &global{
		setup: func() error {
			cert, err := tls.LoadX509KeyPair(cert, key)
			if err != nil {
				return err
			}

			*out = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}

			return nil
		},
	}
}

func HTTPX(notify func(net.Addr, []byte, map[string]interface{}), tlsConfig **tls.Config, srv *httpx.Server) Global {
	return &global{
		setup: func() error {
			wait := sync.Mutex{}

			h := server.HTTPHandler(notify)

			var addr string

			options := []httpx.Option{
				httpx.NotifyStartedFunc(wait.Unlock),
			}

			if tlsConfig == nil {
				addr = ":1080"
			} else {
				addr = ":1443"
				options = append(options, httpx.TLSConfig(*tlsConfig))
			}

			*srv = httpx.New(addr, h, options...)

			go func() {
				if err := (*srv).ListenAndServe(); err != nil {
					log.Fatal(fmt.Errorf("fail to start server: %w", err))
				}
			}()

			if waitTimeout(&wait, 30*time.Second) {
				return errors.New("timeout waiting for server to start")
			}

			return nil
		},
	}
}

func waitTimeout(mu *sync.Mutex, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		mu.Lock()
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}
