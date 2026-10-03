package workerpool

import "golang.org/x/time/rate"

var defaultOptions = options{}

type options struct {
	rateLimit bool
	limit     rate.Limit
	burst     int
}

type Option func(*options)

// RateLimit caps how fast items are handled across all workers.
func RateLimit(limit rate.Limit, burst int) Option {
	return func(opts *options) {
		opts.rateLimit = true
		opts.limit = limit
		opts.burst = burst
	}
}
