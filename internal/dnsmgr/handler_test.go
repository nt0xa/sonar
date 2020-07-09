package dnsmgr_test

import (
	"fmt"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tests = []struct {
	name    string
	qtype   uint16
	results []string
}{
	{"test.sonar.local.", dns.TypeMX, []string{"10 mx.sonar.local"}},
	{"test.sonar.local.", dns.TypeA, []string{"127.0.0.1"}},
	{"test.sonar.local.", dns.TypeAAAA, []string{"127.0.0.1"}},
	{"c1da9f3d.sonar.local.", dns.TypeA, []string{"127.0.0.1"}},
	{"dns1.c1da9f3d.sonar.local.", dns.TypeA, []string{"192.168.1.1", "192.168.1.2"}},
}

func TestDNSMgr(t *testing.T) {
	for _, tt := range tests {
		name := fmt.Sprintf("%s/%s", tt.name, dns.Type(tt.qtype).String())

		t.Run(name, func(st *testing.T) {
			msg := new(dns.Msg)
			msg.Id = dns.Id()
			msg.RecursionDesired = true
			msg.Question = make([]dns.Question, 1)
			msg.Question[0] = dns.Question{
				Name:   tt.name,
				Qtype:  tt.qtype,
				Qclass: dns.ClassINET,
			}

			c := &dns.Client{}
			in, _, err := c.Exchange(msg, "127.0.0.1:1053")
			require.NoError(t, err)
			require.NotNil(t, in)

			require.Len(t, in.Answer, len(tt.results))

			for i, a := range in.Answer {
				assert.Contains(t, a.String(), tt.results[i])
			}
		})
	}
}
