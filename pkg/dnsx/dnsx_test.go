package dnsx_test

import (
	"fmt"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/go-testfixtures/testfixtures"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
			Migrations: "../../internal/database/migrations",
		}, &db),
		testutils.Fixtures(&db, "../../internal/database/fixtures", &tf),
		testutils.DNSX(&db, notifier.Notify, &h, &srv),
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
}

func TestDNS(t *testing.T) {
	for _, tt := range tests {
		tname := fmt.Sprintf("%s/%s", tt.name, dns.Type(tt.qtype).String())

		t.Run(tname, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			name := tt.name

			remoteAddr := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 31337}

			msg := new(dns.Msg)
			msg.Id = dns.Id()
			msg.RecursionDesired = true
			msg.Question = make([]dns.Question, 1)
			msg.Question[0] = dns.Question{
				Name:   tt.name,
				Qtype:  tt.qtype,
				Qclass: dns.ClassINET,
			}

			c := &dns.Client{
				Dialer: &net.Dialer{
					LocalAddr: remoteAddr,
				},
			}

			for i := 0; i < len(tt.results); i++ {

				notifier.
					On("Notify", remoteAddr, mock.MatchedBy(func(data []byte) bool {
						return strings.Contains(string(data), name)
					}),
						map[string]interface{}{
							"Qtype": dns.Type(tt.qtype).String(),
							"Name":  strings.Trim(tt.name, "."),
						}).
					Return()

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
