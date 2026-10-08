package workerpool

import "golang.org/x/time/rate"

var defaultOptions = options{
	workers: 1,
}

type options struct {
	workers   int
	capacity  int
	rateLimit bool
	limit     rate.Limit
	burst     int
}

type Option func(*options)

// WithWorkers sets the number of goroutines running the handler.
func WithWorkers(n int) Option {
	return func(opts *options) {
		opts.workers = n
	}
}

// WithCapacity sets the number of items buffered before Submit blocks.
func WithCapacity(n int) Option {
	return func(opts *options) {
		opts.capacity = n
	}
}

// WithRateLimit caps how fast items are handled across all workers.
func WithRateLimit(limit rate.Limit, burst int) Option {
	return func(opts *options) {
		opts.rateLimit = true
		opts.limit = limit
		opts.burst = burst
	}
}
