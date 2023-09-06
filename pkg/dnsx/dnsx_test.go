package dnsx_test

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/russtone/sonar/pkg/dnsx"
)

var (
	notifier        = &NotifierMock{}
	handlerProvider dnsx.HandlerProvider
)

type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) Notify(remoteAddr net.Addr, data []byte, meta map[string]interface{}) {
	m.Called(remoteAddr, data, meta)
}

func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
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

func TestMain(m *testing.M) {
	var wg sync.WaitGroup
	wg.Add(1)

	// Do not handle DNS queries which are not subdomains of the origin.
	handler := dns.NewServeMux()

	handler.Handle("sonar.test",
		dnsx.NotifyHandler(
			func(e *dnsx.Event) {
				notifier.Notify(e.RemoteAddr, []byte(e.Msg.String()), map[string]interface{}{
					"name":  strings.Trim(e.Msg.Question[0].Name, "."),
					"qtype": dnsx.QtypeString(e.Msg.Question[0].Qtype),
				})
			},
			dnsx.RecordSetHandler(dnsx.NewRecords([]dns.RR{
				dnsx.NewRR("*.sonar.test.", dns.TypeA, 10, "1.1.1.1"),
				dnsx.NewRR("*.sonar.test.", dns.TypeAAAA, 10, "1.1.1.1"),
				dnsx.NewRR("*.sonar.test.", dns.TypeMX, 10, "10 mx.sonar.test."),
				dnsx.NewRR("c1da9f3d.sonar.test.", dns.TypeA, 10, "2.2.2.2"),
			})),
		),
	)

	handlerProvider = dnsx.ChallengeHandler(handler)

	srv := dnsx.New("127.0.0.1:1053", handlerProvider, dnsx.NotifyStartedFunc(wg.Done))

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("fail to start server: %s", err))
			os.Exit(1)
		}
	}()

	if WaitTimeout(&wg, 30*time.Second) {
		fmt.Fprintf(os.Stderr, "timeout waiting for server to start")
		os.Exit(1)
	}

	os.Exit(m.Run())
}

var tests = []struct {
	name    string
	qtype   uint16
	results [][]string
}{
	// Static
	{"test.sonar.test.", dns.TypeMX, [][]string{
		{"10 mx.sonar.test"},
	}},
	{"test.sonar.test.", dns.TypeA, [][]string{
		{"1.1.1.1"},
	}},
	{"test.sonar.test.", dns.TypeAAAA, [][]string{
		{"1.1.1.1"},
	}},
	{"c1da9f3d.sonar.test.", dns.TypeA, [][]string{
		{"2.2.2.2"},
	}},
}

func TestDNS(t *testing.T) {
	for _, tt := range tests {
		tname := fmt.Sprintf("%s/%s", tt.name, dns.Type(tt.qtype).String())

		t.Run(tname, func(t *testing.T) {
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
					On("Notify",
						mock.MatchedBy(func(data net.Addr) bool {
							return data.String() == remoteAddr.String()
						}),
						mock.MatchedBy(func(data []byte) bool {
							return strings.Contains(string(data), name)
						}),
						map[string]interface{}{
							"qtype": dns.Type(tt.qtype).String(),
							"name":  strings.Trim(tt.name, "."),
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

func TestProvider(t *testing.T) {

	for _, name := range []string{
		"_acme-challenge.sonar.test.",
		"_aCme-chAlLEnge.sonar.test.",
	} {

		err := handlerProvider.Present("sonar.test", "", "key1")
		require.NoError(t, err)

		err = handlerProvider.Present("sonar.test", "", "key2")
		require.NoError(t, err)

		msg := new(dns.Msg)
		msg.Id = dns.Id()
		msg.RecursionDesired = true
		msg.Question = make([]dns.Question, 1)
		msg.Question[0] = dns.Question{
			Name:   name,
			Qtype:  dns.TypeTXT,
			Qclass: dns.ClassINET,
		}

		c := &dns.Client{}
		in, _, err := c.Exchange(msg, "127.0.0.1:1053")
		require.NoError(t, err)
		require.Len(t, in.Answer, 2)

		for i, txt := range []string{
			"gXQJloeiZiH04s3XzAOz2s7bP7liJVsar9Azyr6DFTA",
			"sQJTdkyLIz-zdULiNAHHtFDlpvl1HztaAU9vZ-i8mZ0",
		} {
			a, ok := in.Answer[i].(*dns.TXT)
			require.True(t, ok)
			require.Len(t, a.Txt, 1)
			assert.Equal(t, txt, a.Txt[0])
		}

		err = handlerProvider.CleanUp("sonar.test", "", "")
		require.NoError(t, err)

		notifier.On("Notify", mock.Anything, mock.Anything, mock.Anything).Return()

		in, _, err = c.Exchange(msg, "127.0.0.1:1053")
		require.NoError(t, err)
		require.NotNil(t, in)
		require.Len(t, in.Answer, 0)
	}
}
