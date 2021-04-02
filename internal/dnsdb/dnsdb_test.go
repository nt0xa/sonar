package dnsdb_test

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/testutils"
	"github.com/bi-zone/sonar/pkg/dnsrec"
	"github.com/bi-zone/sonar/pkg/dnsx"
)

var (
	db  *database.DB
	tf  *testfixtures.Context
	rec *dnsrec.Records
	h   dnsx.HandlerProvider
	srv dnsx.Server

	notifier = &testutils.NotifierMock{}

	g = testutils.Globals(
		testutils.DB(&database.Config{
			DSN:        os.Getenv("SONAR_DB_DSN"),
			Migrations: "../database/migrations",
		}, &db),
		testutils.Fixtures(&db, "../database/fixtures", &tf),
		testutils.DNSX(&db, notify, &h, &srv),
	)
)

// We don't about notifications here, notifications are tested in dnsx_test.
func notify(net.Addr, []byte, map[string]interface{}) {}

func TestMain(m *testing.M) {
	testutils.TestMain(m, g)
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
	// Static
	{"test.sonar.local.", dns.TypeMX, [][]string{
		{"10 mx.sonar.local"},
	}},
	{"test.sonar.local.", dns.TypeA, [][]string{
		{"127.0.0.1"},
	}},
	{"test.sonar.local.", dns.TypeAAAA, [][]string{
		{"127.0.0.1"},
	}},
	{"c1da9f3d.sonar.local.", dns.TypeA, [][]string{
		{"127.0.0.1"},
	}},

	// Dynamic
	{"test-a.c1da9f3d.sonar.local.", dns.TypeA, [][]string{
		{"192.168.1.1"},
	}},
	{"test-aaaa.c1da9f3d.sonar.local.", dns.TypeAAAA, [][]string{
		{"2001:db8:85a3::8a2e:370:7334"},
	}},
	{"test-mx.c1da9f3d.sonar.local.", dns.TypeMX, [][]string{
		{"10 mx.sonar.local"},
	}},
	{"test-txt.c1da9f3d.sonar.local.", dns.TypeTXT, [][]string{
		{"txt1", "txt2"},
	}},
	{"test-cname.c1da9f3d.sonar.local.", dns.TypeCNAME, [][]string{
		{"example.com"},
	}},
	{"test.test-wildcard.c1da9f3d.sonar.local.", dns.TypeA, [][]string{
		{"192.168.1.1"},
	}},
	{"test2.test-wildcard.c1da9f3d.sonar.local.", dns.TypeA, [][]string{
		{"192.168.1.1"},
	}},

	// Strategies
	{"test-all.c1da9f3d.sonar.local.", dns.TypeA, [][]string{
		{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
	}},
	{"test-round-robin.c1da9f3d.sonar.local.", dns.TypeA, [][]string{
		{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
		{"192.168.1.2", "192.168.1.3", "192.168.1.1"},
		{"192.168.1.3", "192.168.1.1", "192.168.1.2"},
		{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
	}},
	{"test-rebind.c1da9f3d.sonar.local.", dns.TypeA, [][]string{
		{"192.168.1.1"},
		{"192.168.1.2"},
		{"192.168.1.3"},
		{"192.168.1.3"},
	}},

	// No fallback if any custom DNS records added
	{"test.6564e0c7.sonar.local.", dns.TypeA, [][]string{
		{"1.1.1.1"},
	}},
	{"test.6564e0c7.sonar.local.", dns.TypeAAAA, [][]string{
		{},
	}},
	{"test.6564e0c7.sonar.local.", dns.TypeMX, [][]string{
		{},
	}},
}

func TestDNS(t *testing.T) {
	for _, tt := range tests {
		tname := fmt.Sprintf("%s/%s", tt.name, dns.Type(tt.qtype).String())

		t.Run(tname, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			msg := new(dns.Msg)
			msg.Id = dns.Id()
			msg.RecursionDesired = true
			msg.Question = make([]dns.Question, 1)
			msg.Question[0] = dns.Question{
				Name:   tt.name,
				Qtype:  tt.qtype,
				Qclass: dns.ClassINET,
			}

			c := new(dns.Client)

			for i := 0; i < len(tt.results); i++ {

				in, _, err := c.Exchange(msg, "127.0.0.1:1053")
				require.NoError(t, err)
				require.NotNil(t, in)

				require.Len(t, in.Answer, len(tt.results[i]))

				for j, a := range in.Answer {
					assert.Contains(t, a.String(), tt.results[i][j])
					assert.Equal(t, tt.name, a.Header().Name)
				}
			}
		})
	}

	notifier.AssertExpectations(t)
}
