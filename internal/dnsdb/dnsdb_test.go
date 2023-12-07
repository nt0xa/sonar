package dnsdb_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/russtone/sonar/internal/database"
	"github.com/russtone/sonar/internal/dnsdb"
)

var (
	tf  *testfixtures.Loader
	db  *database.DB
	rec *dnsdb.Records
)

func TestMain(m *testing.M) {
	var (
		dsn string
		err error
	)

	if dsn = os.Getenv("SONAR_DB_DSN"); dsn == "" {
		fmt.Fprintln(os.Stderr, "empty SONAR_DB_DSN")
		os.Exit(1)
	}

	db, err = database.New(&database.Config{
		DSN:        dsn,
		Migrations: "../database/migrations",
	}, logrus.New())
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to init database: %v\n", err)
		os.Exit(1)
	}

	if err := db.Migrate(); err != nil {
		fmt.Fprintf(os.Stderr, "fail to apply database migrations: %v\n", err)
		os.Exit(1)
	}

	rec = &dnsdb.Records{DB: db, Origin: "sonar.test"}

	tf, err = testfixtures.New(
		testfixtures.Database(db.DB.DB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("../database/fixtures"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to load fixtures: %v", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func setup(t *testing.T) {
	err := tf.Load()
	require.NoError(t, err)
}

func teardown(t *testing.T) {}

var tests = []struct {
	name    string
	qtype   uint16
	results [][]string
}{
	// Dynamic
	{"test-a.c1da9f3d.sonar.test.", dns.TypeA, [][]string{
		{"192.168.1.1"},
	}},
	{"test-aaaa.c1da9f3d.sonar.test.", dns.TypeAAAA, [][]string{
		{"2001:db8:85a3::8a2e:370:7334"},
	}},
	{"test-mx.c1da9f3d.sonar.test.", dns.TypeMX, [][]string{
		{"10 mx.sonar.test"},
	}},
	{"test-txt.c1da9f3d.sonar.test.", dns.TypeTXT, [][]string{
		{"txt1", "txt2"},
	}},
	{"test-cname.c1da9f3d.sonar.test.", dns.TypeCNAME, [][]string{
		{"example.com"},
	}},
	{"test.test-wildcard.c1da9f3d.sonar.test.", dns.TypeA, [][]string{
		{"192.168.1.1"},
	}},
	{"test2.test-wildcard.c1da9f3d.sonar.test.", dns.TypeA, [][]string{
		{"192.168.1.1"},
	}},
	{"test2.test-wildcard.c1da9f3d.sonar.test.", dns.TypeA, [][]string{
		{"192.168.1.1"},
	}},
	{"test-ns.c1da9f3d.sonar.test.", dns.TypeNS, [][]string{
		{"ns1.example.com."},
	}},

	// Strategies
	{"test-all.c1da9f3d.sonar.test.", dns.TypeA, [][]string{
		{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
	}},
	{"test-round-robin.c1da9f3d.sonar.test.", dns.TypeA, [][]string{
		{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
		{"192.168.1.2", "192.168.1.3", "192.168.1.1"},
		{"192.168.1.3", "192.168.1.1", "192.168.1.2"},
		{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
	}},
	{"test-rebind.c1da9f3d.sonar.test.", dns.TypeA, [][]string{
		{"192.168.1.1"},
		{"192.168.1.2"},
		{"192.168.1.3"},
		{"192.168.1.3"},
	}},

	// No fallback if any custom DNS records added
	{"test.6564e0c7.sonar.test.", dns.TypeA, [][]string{
		{"1.1.1.1"},
	}},
	{"test.6564e0c7.sonar.test.", dns.TypeAAAA, [][]string{
		{},
	}},
	{"test.6564e0c7.sonar.test.", dns.TypeMX, [][]string{
		{},
	}},
}

func TestDNS(t *testing.T) {
	for _, tt := range tests {
		tname := fmt.Sprintf("%s/%s", tt.name, dns.Type(tt.qtype).String())

		t.Run(tname, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			for i := 0; i < len(tt.results); i++ {

				rrs, err := rec.Get(tt.name, tt.qtype)
				require.NoError(t, err)

				require.Len(t, rrs, len(tt.results[i]))

				for j, rr := range rrs {
					assert.Contains(t, rr.String(), tt.results[i][j])
					assert.Equal(t, tt.name, rr.Header().Name)
				}
			}
		})
	}
}
