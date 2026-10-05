// Package coalescer batches items by key with a growing window.
package coalescer

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nt0xa/sonar/pkg/workerpool"
)

// Coalescer groups items by key. The first passThrough items for a key are
// emitted immediately, subsequent items are buffered and emitted as a batch
// when the key's window expires. Every window that emits a batch doubles the next
// one (up to maxWindow); a window with nothing buffered resets the key.
//
// Batches are handled by the handler on worker goroutines, outside the
// coalescer's lock, so the handler may block or call Push. It must not call
// Stop: Stop waits for all workers, including the one running the handler.
type Coalescer[T any] struct {
	keyFn KeyFn[T]
	opts  options
	pool  *workerpool.Pool[[]Item[T]]

	mu      sync.Mutex
	entries map[string]*entry[T]
	stopped bool

	dropped atomic.Int64
}

type KeyFn[T any] = func(T) string

// Item is a pushed value together with the context it was pushed with.
type Item[T any] struct {
	Ctx   context.Context
	Value T
}

type entry[T any] struct {
	buf    []Item[T]
	window time.Duration
	count  int
	timer  *time.Timer
}

// New creates a Coalescer, panics on invalid arguments.
func New[T any](keyFn KeyFn[T], handler func([]Item[T]), opts ...Option) *Coalescer[T] {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	if keyFn == nil {
		panic("coalescer: keyFn must not be nil")
	}

	if handler == nil {
		panic("coalescer: handler must not be nil")
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

	if options.workers < 1 {
		panic("coalescer: workers must be >= 1")
	}

	options.maxWindow = max(options.maxWindow, options.window)

	return &Coalescer[T]{
		keyFn: keyFn,
		opts:  options,
		pool: workerpool.New(options.workers, options.bufferSize,
			func(_ context.Context, b []Item[T]) { handler(b) },
		),
		entries: make(map[string]*entry[T]),
	}
}

// Push adds an item, never blocks; overflow is counted in Dropped. The
// caller's ctx is preserved for its trace span and values but stripped of
// cancellation/deadline, since the item may be handled after ctx ends.
func (c *Coalescer[T]) Push(ctx context.Context, value T) {
	key := c.keyFn(value)
	item := Item[T]{Ctx: context.WithoutCancel(ctx), Value: value}

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
			timer:  time.AfterFunc(c.opts.window, func() { c.tick(key) }),
		}
		c.emit([]Item[T]{item})
		return
	}

	if e.count < c.opts.passThrough {
		c.emit([]Item[T]{item})
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
	e.timer = time.AfterFunc(e.window, func() { c.tick(key) })
}

// emit must be called with c.mu held and !c.stopped, so it never sends on
// the pool after Stop has closed it.
func (c *Coalescer[T]) emit(batch []Item[T]) {
	if !c.pool.TryProcess(context.Background(), batch) {
		c.dropped.Add(int64(len(batch)))
	}
}

// Dropped returns the number of items dropped because a key's batch was
// full (maxBatch) or the handler queue was full (bufferSize).
func (c *Coalescer[T]) Dropped() int64 {
	return c.dropped.Load()
}

// Stop hands all pending batches to the handler queue and waits until the
// handlers finish or ctx is done. A non-nil error means shutdown was started
// but handlers may still be running; call Stop again to keep waiting. Items
// pushed after Stop are ignored. Stop is safe to call multiple times but must
// not be called from the handler.
func (c *Coalescer[T]) Stop(ctx context.Context) error {
	c.mu.Lock()
	if !c.stopped {
		for _, e := range c.entries {
			e.timer.Stop()
			if len(e.buf) != 0 {
				c.emit(e.buf)
			}
		}
		clear(c.entries)
		c.stopped = true
	}
	c.mu.Unlock()

	return c.pool.Stop(ctx)
}
