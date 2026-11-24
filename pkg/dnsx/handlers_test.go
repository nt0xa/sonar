package dnsx_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/pkg/dnsx"
)

// responseWriter is a dns.ResponseWriter that discards responses.
type responseWriter struct {
	dns.ResponseWriter
}

func (responseWriter) RemoteAddr() net.Addr    { return &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)} }
func (responseWriter) WriteMsg(*dns.Msg) error { return nil }

func TestNotifyHandlerAnswer(t *testing.T) {
	var meta *dnsx.Meta

	h := dnsx.NotifyHandler(
		func(_ context.Context, _ net.Addr, _ *time.Time, _ [][]byte, _ []byte, m *dnsx.Meta) { meta = m },
		dnsx.RecordSetHandler(dnsx.NewRecords([]dns.RR{
			dnsx.NewRR("test.sonar.test.", dns.TypeA, 10, "1.1.1.1"),
		})),
	)

	h.ServeDNS(context.Background(), responseWriter{}, new(dns.Msg).SetQuestion("test.sonar.test.", dns.TypeA))

	require.NotNil(t, meta)
	assert.Equal(t, []dnsx.Answer{{Name: "test.sonar.test", Type: "A", TTL: 10}}, meta.Answer)
}
