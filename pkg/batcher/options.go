package batcher

import (
	"time"
)

var defaultOptions = options{
	window:         time.Second,
	maxWindow:      time.Minute,
	maxBatch:       100,
	outputCapacity: 100,
	passThrough:    1,
}

type options struct {
	window         time.Duration
	maxWindow      time.Duration
	maxBatch       int
	outputCapacity int
	passThrough    int
}

type Option func(*options)

// WithWindow sets the initial window for a key.
func WithWindow(d time.Duration) Option {
	return func(opts *options) {
		opts.window = d
	}
}

// WithMaxWindow sets the upper bound for the window growth.
// It is raised to WithWindow if smaller.
func WithMaxWindow(d time.Duration) Option {
	return func(opts *options) {
		opts.maxWindow = d
	}
}

// WithMaxBatch sets the maximum number of items buffered per key within a window.
// Items beyond it are dropped.
func WithMaxBatch(n int) Option {
	return func(opts *options) {
		opts.maxBatch = n
	}
}

// WithOutputCapacity sets the capacity of the Batches channel.
// Batches are dropped when it is full.
func WithOutputCapacity(n int) Option {
	return func(opts *options) {
		opts.outputCapacity = n
	}
}

// WithPassThrough sets the number of items per key that are emitted immediately
// before batching is applied. The count resets after a window in which
// nothing was buffered, so keys with at most n items per window are never
// batched. Must be >= 1.
func WithPassThrough(n int) Option {
	return func(opts *options) {
		opts.passThrough = n
	}
}
