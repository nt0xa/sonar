package dns_test

import (
	"net"
	"strings"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var tests = []struct {
	name         string
	qtype        uint16
	meta         map[string]interface{}
	dataContains string
}{
	{"test.sonar.local.", dns.TypeMX, map[string]interface{}{"Qtype": "MX"}, "test"},
	{"test.sonar.local.", dns.TypeA, map[string]interface{}{"Qtype": "A"}, "test"},
	{"test.sonar.local.", dns.TypeAAAA, map[string]interface{}{"Qtype": "AAAA"}, "test"},
}

func TestDNS(t *testing.T) {
	for _, tt := range tests {
		name := dns.Type(tt.qtype).String()

		t.Run(name, func(st *testing.T) {

			remoteAddr := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 31337}

			notifier.
				On("Notify", remoteAddr, mock.MatchedBy(func(data []byte) bool {
					return strings.Contains(string(data), tt.dataContains)
				}), tt.meta).
				Return()

			handler.
				On("HandleFunc", mock.Anything, mock.Anything).
				Return()

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
			in, _, err := c.Exchange(msg, "127.0.0.1:1053")
			require.NoError(t, err)
			require.NotNil(t, in)
		})
	}
}
