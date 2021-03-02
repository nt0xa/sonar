package testutils

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/database/dbactions"
	"github.com/bi-zone/sonar/internal/modules/api"
	"github.com/bi-zone/sonar/internal/modules/api/apiclient"
	"github.com/bi-zone/sonar/internal/protocols/dnsx"
	"github.com/bi-zone/sonar/internal/protocols/dnsx/dnsdb"
	"github.com/bi-zone/sonar/internal/protocols/dnsx/dnsdef"
	"github.com/bi-zone/sonar/internal/protocols/dnsx/dnsrec"
	"github.com/bi-zone/sonar/internal/utils/logger"
	"github.com/go-testfixtures/testfixtures"
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
			*out = dbactions.New(*db, log, TestDomain)
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

func DNSDefaultRecords(out **dnsrec.Records) Global {
	return &global{
		setup: func() error {
			rec, err := dnsdef.Records(TestDomain, TestIP)
			if err != nil {
				return fmt.Errorf("fail to init default dns records: %w", err)
			}

			*out = rec
			return nil
		},
	}
}

func DNSDBRecords(db **database.DB, out **dnsdb.Handler) Global {
	return &global{
		setup: func() error {
			*out = &dnsdb.Handler{
				DB:     *db,
				Origin: TestDomain,
			}
			return nil
		},
	}
}

func DNSX(handlers [](func() dnsx.Handler), notify func(net.Addr, []byte, map[string]interface{}), out **dnsx.Server) Global {
	return &global{
		setup: func() error {
			wg := sync.WaitGroup{}
			wg.Add(1)

			hh := make([]dnsx.Handler, 0)

			for _, h := range handlers {
				hh = append(hh, h())
			}

			*out = &dnsx.Server{
				Addr:     ":1053",
				Origin:   TestDomain,
				Handlers: hh,
				NotifyStartedFunc: func() {
					wg.Done()
				},
				NotifyRequestFunc: notify,
			}

			go func() {
				if err := (*out).ListenAndServe(); err != nil {
					log.Fatal(fmt.Errorf("fail to start server: %w", err))
				}
			}()

			if WaitTimeout(&wg, 30*time.Second) {
				return errors.New("timeout waiting for server to start")
			}

			return nil
		},
		teardown: func() error {
			return (*out).Shutdown()
		},
	}
}
