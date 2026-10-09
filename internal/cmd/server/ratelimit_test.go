package server

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/pkg/ftpx"
	"github.com/nt0xa/sonar/pkg/netx"
	"github.com/nt0xa/sonar/pkg/smtpx"
	"github.com/nt0xa/sonar/pkg/telemetry"
)

func Test_RateLimitAllowFunc_Disabled(t *testing.T) {
	addr := net.TCPAddrFromAddrPort(netip.MustParseAddrPort("192.0.2.1:1000"))

	for name, allow := range map[string]func(net.Addr) bool{
		"not enabled": RateLimitAllowFunc(&RateLimitConfig{Enabled: false}, RateLimitProtoConfig{Rate: 1, Burst: 1}, telemetry.NewNoop(), "http"),
		"zero rate":   RateLimitAllowFunc(&RateLimitConfig{Enabled: true}, RateLimitProtoConfig{}, telemetry.NewNoop(), "http"),
	} {
		t.Run(name, func(t *testing.T) {
			for range 10 {
				require.True(t, allow(addr))
			}
		})
	}
}

func Test_RateLimitAllowFunc(t *testing.T) {
	allow := RateLimitAllowFunc(
		&RateLimitConfig{Enabled: true, Allow: []string{"10.0.0.0/8"}},
		RateLimitProtoConfig{Rate: 0.001, Burst: 1},
		telemetry.NewNoop(),
		"http",
	)
	addr := func(s string) net.Addr {
		return net.TCPAddrFromAddrPort(netip.MustParseAddrPort(s))
	}

	assert.True(t, allow(addr("192.0.2.1:1000")))
	assert.False(t, allow(addr("192.0.2.1:1000")))

	assert.True(t, allow(addr("10.1.2.3:1000")))
	assert.True(t, allow(addr("10.1.2.3:1000")))
}

func Test_SessionHandlersDropRateLimitedConnections(t *testing.T) {
	log := slog.New(slog.DiscardHandler)
	deny := func(net.Addr) bool { return false }

	handlers := map[string]netx.Handler{
		"smtp": SMTPHandler("example.com", log, telemetry.NewNoop(), nil, deny,
			func(context.Context, net.Addr, *time.Time, bool, [][]byte, []byte, *smtpx.Meta) {
				t.Error("unexpected SMTP event")
			}),
		"ftp": FTPHandler("example.com", log, telemetry.NewNoop(), deny,
			func(context.Context, net.Addr, *time.Time, bool, [][]byte, []byte, *ftpx.Meta) {
				t.Error("unexpected FTP event")
			}),
	}

	for name, h := range handlers {
		t.Run(name, func(t *testing.T) {
			srv, cli := net.Pipe()

			read := make(chan []byte)
			go func() {
				data, _ := io.ReadAll(cli)
				read <- data
			}()

			h.Handle(t.Context(), srv)
			_ = srv.Close()

			assert.Empty(t, <-read, "no greeting must be sent")
		})
	}
}
