package dnsx_test

import (
	"net"
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/protocols/dnsx"
	"github.com/bi-zone/sonar/internal/protocols/dnsx/dnsdb"
	"github.com/bi-zone/sonar/internal/protocols/dnsx/dnsrec"
	"github.com/bi-zone/sonar/internal/testutils"
)

var (
	db    *database.DB
	tf    *testfixtures.Context
	rec   *dnsrec.Records
	dbrec *dnsdb.Handler
	srv   *dnsx.Server

	notifier = &NotifierMock{}

	g = testutils.Globals(
		testutils.DB(&database.Config{
			DSN:        os.Getenv("SONAR_DB_DSN"),
			Migrations: "../database/migrations",
		}, &db),
		testutils.Fixtures(&db, "../database/fixtures", &tf),
		testutils.DNSDefaultRecords(&rec),
		testutils.DNSDBRecords(&db, &dbrec),
		testutils.DNSX([](func() dnsx.Handler){
			func(r **dnsdb.Handler) func() dnsx.Handler {
				return func() dnsx.Handler {
					return *r
				}
			}(&dbrec),
			func(r **dnsrec.Records) func() dnsx.Handler {
				return func() dnsx.Handler {
					return *r
				}
			}(&rec),
		}, notifier.Notify, &srv),
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

type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) Notify(remoteAddr net.Addr, data []byte, meta map[string]interface{}) {
	m.Called(remoteAddr, data, meta)
}
