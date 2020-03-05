package dns_test

import (
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dnssrv "github.com/bi-zone/sonar/pkg/server/dns"
	"github.com/bi-zone/sonar/pkg/server/mock_server"
)

var tests = []struct {
	name            string
	qtype           uint16
	meta            map[string]interface{}
	dataContains    string
	answersContains []string
}{
	{"test.sonar.local", dns.TypeMX, map[string]interface{}{"Qtype": "MX"}, "test", []string{"mx.sonar.local"}},
	{"test.sonar.local", dns.TypeA, map[string]interface{}{"Qtype": "A"}, "test", []string{"127.0.0.1"}},
	{"test.sonar.local", dns.TypeAAAA, map[string]interface{}{"Qtype": "AAAA"}, "test", []string{"127.0.0.1"}},
	{"test.sonar.local", dns.TypeNS, map[string]interface{}{"Qtype": "NS"}, "test", []string{"ns1.sonar.local", "127.0.0.1"}},
	{"test.sonar.local", dns.TypeSOA, map[string]interface{}{"Qtype": "SOA"}, "test", []string{"ns1.sonar.local"}},
}

func TestDNS(t *testing.T) {
	for _, tt := range tests {
		name := dns.Type(tt.qtype).String()

		t.Run(name, func(st *testing.T) {
			ctrl := gomock.NewController(st)
			defer ctrl.Finish()

			m := mock_server.NewMockRequestNotifier(ctrl)
			srv.SetOption(dnssrv.NotifyRequestFunc(m.Notify))

			remoteAddr := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 31337}

			m.
				EXPECT().
				Notify(gomock.Eq(remoteAddr), Contains(tt.dataContains), gomock.Eq(tt.meta)).
				Times(1)

			msg := new(dns.Msg)
			msg.Id = dns.Id()
			msg.RecursionDesired = true
			msg.Question = make([]dns.Question, 1)
			msg.Question[0] = dns.Question{
				Name:   "test.sonar.local.",
				Qtype:  tt.qtype,
				Qclass: dns.ClassINET,
			}

			c := &dns.Client{
				Dialer: &net.Dialer{
					LocalAddr: remoteAddr,
				},
			}
			in, _, err := c.Exchange(msg, "127.0.0.1:1053")
			require.NoError(t, err)
			require.NotNil(t, in)
			assert.Len(t, in.Answer, len(tt.answersContains))
			for i, s := range in.Answer {
				assert.Contains(t, s.String(), tt.answersContains[i])
			}
		})
	}
}

func TestDNS_TXT(t *testing.T) {
	srv.SetOption(
		dnssrv.NotifyRequestFunc(func(net.Addr, []byte, map[string]interface{}) {}),
	)

	err := srv.Present("sonar.local", "", "key1")
	require.NoError(t, err)

	err = srv.Present("sonar.local", "", "key2")
	require.NoError(t, err)

	msg := new(dns.Msg)
	msg.Id = dns.Id()
	msg.RecursionDesired = true
	msg.Question = make([]dns.Question, 1)
	msg.Question[0] = dns.Question{
		Name:   "_acme-challenge.sonar.local.",
		Qtype:  dns.TypeTXT,
		Qclass: dns.ClassINET,
	}

	c := &dns.Client{}
	in, _, err := c.Exchange(msg, "127.0.0.1:1053")
	require.NoError(t, err)
	require.NotNil(t, in)
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

	err = srv.CleanUp("sonar.local", "", "")
	require.NoError(t, err)

	in, _, err = c.Exchange(msg, "127.0.0.1:1053")
	require.NoError(t, err)
	require.NotNil(t, in)
	require.Len(t, in.Answer, 0)
}
