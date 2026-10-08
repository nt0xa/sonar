package ratelimit

import "net/netip"

var defaultOptions = options{
	maxEntries: 100_000,
}

type options struct {
	allow      []netip.Prefix
	maxEntries int
}

type Option func(*options)

// WithAllow exempts addresses within the given prefixes from limiting.
func WithAllow(prefixes ...netip.Prefix) Option {
	return func(opts *options) {
		opts.allow = append(opts.allow, prefixes...)
	}
}

// WithMaxEntries caps the number of tracked keys; new keys beyond it are not limited.
func WithMaxEntries(n int) Option {
	return func(opts *options) {
		opts.maxEntries = n
	}
}
