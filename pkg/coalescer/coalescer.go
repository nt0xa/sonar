// Package coalescer batches items by key with a growing window.
package coalescer

import (
	"context"
	"iter"
	"sync"
	"sync/atomic"
	"time"
)

// Coalescer groups items by key. The first passThrough items for a key are
// emitted immediately, subsequent items are buffered and emitted as a batch
// when the key's window expires. Every window that emits a batch doubles the next
// one (up to maxWindow); a window with nothing buffered resets the key.
type Coalescer[T any] struct {
	keyFn KeyFn[T]
	opts  options

	mu      sync.Mutex
	entries map[string]*entry[T]
	out     chan []T
	stopped bool

	dropped atomic.Int64
}

type KeyFn[T any] = func(T) string

type entry[T any] struct {
	buf    []T
	window time.Duration
	count  int
}

// New creates a Coalescer, panics on invalid arguments.
func New[T any](keyFn KeyFn[T], opts ...Option) *Coalescer[T] {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	if keyFn == nil {
		panic("coalescer: keyFn must not be nil")
	}

	if options.window <= 0 {
		panic("coalescer: window must be > 0")
	}

	if options.maxBatch < 0 {
		panic("coalescer: maxBatch must be >= 0")
	}

	if options.bufferSize < 0 {
		panic("coalescer: bufferSize must be >= 0")
	}

	if options.passThrough < 1 {
		panic("coalescer: passThrough must be >= 1")
	}

	options.maxWindow = max(options.maxWindow, options.window)

	return &Coalescer[T]{
		keyFn:   keyFn,
		opts:    options,
		entries: make(map[string]*entry[T]),
		out:     make(chan []T, options.bufferSize),
	}
}

// Push adds an item, never blocks; overflow is counted in Dropped.
func (c *Coalescer[T]) Push(item T) {
	key := c.keyFn(item)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopped {
		return
	}

	e, ok := c.entries[key]
	if !ok {
		c.entries[key] = &entry[T]{
			window: c.opts.window,
			count:  1,
		}
		c.emit([]T{item})
		time.AfterFunc(c.opts.window, func() { c.tick(key) })
		return
	}

	if e.count < c.opts.passThrough {
		c.emit([]T{item})
		e.count++
		return
	}

	if len(e.buf) >= c.opts.maxBatch {
		c.dropped.Add(1)
		return
	}

	e.buf = append(e.buf, item)
}

func (c *Coalescer[T]) tick(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopped {
		return
	}

	e := c.entries[key]

	if len(e.buf) == 0 {
		delete(c.entries, key)
		return
	}

	c.emit(e.buf)
	e.buf = nil
	e.window = min(e.window*2, c.opts.maxWindow)
	time.AfterFunc(e.window, func() { c.tick(key) })
}

func (c *Coalescer[T]) emit(batch []T) {
	select {
	case c.out <- batch:
	default:
		c.dropped.Add(int64(len(batch)))
	}
}

// Dropped returns the number of items dropped because a key's batch was
// full or the output buffer was full.
func (c *Coalescer[T]) Dropped() int64 {
	return c.dropped.Load()
}

// Next yields batches until ctx is done or Stop is called.
func (c *Coalescer[T]) Next(ctx context.Context) iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		for {
			select {
			case <-ctx.Done():
				return
			case it, ok := <-c.out:
				if !ok {
					return
				}
				if !yield(it) {
					return
				}
			}
		}
	}
}

// Stop emits all pending batches and closes the output. Items pushed
// after Stop are ignored. Stop is safe to call multiple times.
func (c *Coalescer[T]) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopped {
		return
	}
	c.stopped = true

	for _, e := range c.entries {
		if len(e.buf) != 0 {
			c.emit(e.buf)
		}
	}
	clear(c.entries)

	close(c.out)
}
