package ratelimit_test

import (
	"net"
	"net/netip"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/nt0xa/sonar/pkg/ratelimit"
)

func tcp(s string) net.Addr {
	return net.TCPAddrFromAddrPort(netip.MustParseAddrPort(s))
}

func udp(s string) net.Addr {
	return net.UDPAddrFromAddrPort(netip.MustParseAddrPort(s))
}

// drain consumes the whole burst for addr.
func drain(t *testing.T, l *ratelimit.Limiter, addr net.Addr, burst int) {
	t.Helper()

	for i := range burst {
		require.True(t, l.Allow(addr), "request %d", i)
	}

	require.False(t, l.Allow(addr))
}

func Test_Panics(t *testing.T) {
	require.Panics(t, func() { ratelimit.New(0, 1) })
	require.Panics(t, func() { ratelimit.New(1, 0) })
	require.Panics(t, func() { ratelimit.New(1, 1, ratelimit.WithMaxEntries(0)) })
}

func Test_LimitsAfterBurstAndRefills(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := ratelimit.New(1, 3)
		addr := tcp("192.0.2.1:1234")

		drain(t, l, addr, 3)

		time.Sleep(time.Second)

		assert.True(t, l.Allow(addr))
		assert.False(t, l.Allow(addr))
	})
}

func Test_KeysByIPv4AddressIgnoringPort(t *testing.T) {
	l := ratelimit.New(rate.Every(time.Hour), 2)

	drain(t, l, tcp("192.0.2.1:1000"), 2)

	assert.False(t, l.Allow(udp("192.0.2.1:2000")))
	assert.True(t, l.Allow(tcp("192.0.2.2:1000")))
}

func Test_KeysByIPv6Slash64(t *testing.T) {
	l := ratelimit.New(rate.Every(time.Hour), 2)

	drain(t, l, tcp("[2001:db8:0:1::1]:1000"), 2)

	assert.False(t, l.Allow(tcp("[2001:db8:0:1:ffff::2]:1000")))
	assert.True(t, l.Allow(tcp("[2001:db8:0:2::1]:1000")))
}

func Test_IPv4MappedIPv6SharesIPv4Bucket(t *testing.T) {
	l := ratelimit.New(rate.Every(time.Hour), 2)

	drain(t, l, tcp("192.0.2.1:1000"), 2)

	assert.False(t, l.Allow(tcp("[::ffff:192.0.2.1]:1000")))
}

func Test_AllowlistIsNeverLimited(t *testing.T) {
	l := ratelimit.New(rate.Every(time.Hour), 1,
		ratelimit.WithAllow(netip.MustParsePrefix("10.0.0.0/8")),
	)

	for range 10 {
		require.True(t, l.Allow(tcp("10.1.2.3:1000")))
	}

	drain(t, l, tcp("192.0.2.1:1000"), 1)
}

func Test_NonIPAddrIsAllowed(t *testing.T) {
	l := ratelimit.New(rate.Every(time.Hour), 1)
	addr := &net.UnixAddr{Name: "/tmp/sock", Net: "unix"}

	for range 10 {
		require.True(t, l.Allow(addr))
	}
}

func Test_FailsOpenAtMaxEntries(t *testing.T) {
	l := ratelimit.New(rate.Every(time.Hour), 1, ratelimit.WithMaxEntries(1))

	drain(t, l, tcp("192.0.2.1:1000"), 1)

	// Untracked keys are not limited, tracked ones still are.
	for range 10 {
		require.True(t, l.Allow(tcp("192.0.2.2:1000")))
	}

	assert.False(t, l.Allow(tcp("192.0.2.1:1000")))
}

func Test_SweepDropsRefilledBuckets(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := ratelimit.New(1, 1, ratelimit.WithMaxEntries(1))

		drain(t, l, tcp("192.0.2.1:1000"), 1)

		time.Sleep(time.Minute)

		// The refilled bucket was swept, so the new key gets tracked and limited.
		drain(t, l, tcp("192.0.2.2:1000"), 1)
	})
}

func Test_SweepKeepsDrainedBuckets(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := ratelimit.New(rate.Every(time.Hour), 1)
		addr := tcp("192.0.2.1:1000")

		drain(t, l, addr, 1)

		time.Sleep(time.Minute)

		// Triggers a sweep that must not reset the drained bucket.
		assert.False(t, l.Allow(addr))
	})
}
