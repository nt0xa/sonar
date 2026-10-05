package coalescer

import (
	"time"
)

var defaultOptions = options{
	window:      time.Second,
	maxWindow:   time.Minute,
	maxBatch:    100,
	bufferSize:  100,
	passThrough: 1,
	workers:     1,
}

type options struct {
	window      time.Duration
	maxWindow   time.Duration
	maxBatch    int
	bufferSize  int
	passThrough int
	workers     int
}

type Option func(*options)

// Window sets the initial window for a key.
func Window(d time.Duration) Option {
	return func(opts *options) {
		opts.window = d
	}
}

// MaxWindow sets the upper bound for the window growth.
// It is raised to Window if smaller.
func MaxWindow(d time.Duration) Option {
	return func(opts *options) {
		opts.maxWindow = d
	}
}

// MaxBatch sets the maximum number of items buffered per key within a window.
// Items beyond it are dropped.
func MaxBatch(n int) Option {
	return func(opts *options) {
		opts.maxBatch = n
	}
}

// BufferSize sets the capacity of the handler queue.
// Batches are dropped when it is full.
func BufferSize(n int) Option {
	return func(opts *options) {
		opts.bufferSize = n
	}
}

// Workers sets the number of goroutines running the handler.
func Workers(n int) Option {
	return func(opts *options) {
		opts.workers = n
	}
}

// PassThrough sets the number of items per key that are emitted immediately
// before coalescing is applied. The count resets after a window in which
// nothing was buffered, so keys with at most n items per window are never
// coalesced. Must be >= 1.
func PassThrough(n int) Option {
	return func(opts *options) {
		opts.passThrough = n
	}
}
