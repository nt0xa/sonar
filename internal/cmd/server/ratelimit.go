package server

import (
	"context"
	"net"
	"net/netip"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"golang.org/x/time/rate"

	"github.com/nt0xa/sonar/pkg/netx"
	"github.com/nt0xa/sonar/pkg/ratelimit"
	"github.com/nt0xa/sonar/pkg/telemetry"
)

// RateLimitAllowFunc returns a per-IP limit check that counts rejections; it allows everything when disabled.
func RateLimitAllowFunc(
	cfg *RateLimitConfig,
	proto RateLimitProtoConfig,
	tel telemetry.Telemetry,
	protocol string,
) func(net.Addr) bool {
	if !cfg.Enabled || proto.Rate == 0 {
		return func(net.Addr) bool { return true }
	}

	prefixes := make([]netip.Prefix, 0, len(cfg.Allow))
	for _, s := range cfg.Allow {
		// Already validated by config.
		prefixes = append(prefixes, netip.MustParsePrefix(s))
	}

	l := ratelimit.New(rate.Limit(proto.Rate), proto.Burst, ratelimit.WithAllow(prefixes...))

	rejected, err := tel.NewInt64Counter(
		"ratelimit.rejected",
		"{count}",
		"Number of requests rejected by per-IP rate limiting",
	)
	if err != nil {
		panic(err)
	}

	attrs := metric.WithAttributes(attribute.String("protocol", protocol))

	return func(addr net.Addr) bool {
		if l.Allow(addr) {
			return true
		}

		rejected.Add(context.Background(), 1, attrs)

		return false
	}
}

// RateLimitListenerWrapper closes accepted connections over the limit.
func RateLimitListenerWrapper(allow func(net.Addr) bool) func(net.Listener) net.Listener {
	return func(l net.Listener) net.Listener {
		return &netx.RateLimitListener{Listener: l, Allow: allow}
	}
}
