// Package ratelimit provides a per-IP token bucket rate limiter keyed by /32
// for IPv4 and /64 for IPv6.
package ratelimit

import (
	"net"
	"net/netip"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// sweepInterval is how often buckets that have fully refilled are dropped.
const sweepInterval = time.Minute

type Limiter struct {
	limit      rate.Limit
	burst      int
	allow      []netip.Prefix
	maxEntries int

	// mu guards entries and lastSweep.
	mu        sync.Mutex
	entries   map[netip.Prefix]*rate.Limiter
	lastSweep time.Time
}

// New creates a Limiter, panics on invalid arguments.
func New(limit rate.Limit, burst int, opts ...Option) *Limiter {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	if limit <= 0 {
		panic("ratelimit: limit must be > 0")
	}

	if burst <= 0 {
		panic("ratelimit: burst must be > 0")
	}

	if options.maxEntries <= 0 {
		panic("ratelimit: max entries must be > 0")
	}

	return &Limiter{
		limit:      limit,
		burst:      burst,
		allow:      options.allow,
		maxEntries: options.maxEntries,
		entries:    make(map[netip.Prefix]*rate.Limiter),
		lastSweep:  time.Now(),
	}
}

// Allow reports whether a request from addr may proceed; non-IP addresses are always allowed.
func (l *Limiter) Allow(addr net.Addr) bool {
	ip, ok := addrIP(addr)
	if !ok {
		return true
	}

	for _, p := range l.allow {
		if p.Contains(ip) {
			return true
		}
	}

	key := keyOf(ip)
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastSweep) >= sweepInterval {
		l.sweep(now)
	}

	lim, ok := l.entries[key]
	if !ok {
		// Fail open rather than drop traffic from untracked sources.
		if len(l.entries) >= l.maxEntries {
			return true
		}

		lim = rate.NewLimiter(l.limit, l.burst)
		l.entries[key] = lim
	}

	return lim.AllowN(now, 1)
}

// sweep drops full buckets, which behave the same as missing ones.
func (l *Limiter) sweep(now time.Time) {
	for k, lim := range l.entries {
		if lim.TokensAt(now) >= float64(l.burst) {
			delete(l.entries, k)
		}
	}

	l.lastSweep = now
}

func addrIP(addr net.Addr) (netip.Addr, bool) {
	a, ok := addr.(interface{ AddrPort() netip.AddrPort })
	if !ok {
		return netip.Addr{}, false
	}

	ip := a.AddrPort().Addr().Unmap().WithZone("")

	return ip, ip.IsValid()
}

func keyOf(ip netip.Addr) netip.Prefix {
	bits := 32
	if ip.Is6() {
		bits = 64
	}

	p, _ := ip.Prefix(bits)

	return p
}
