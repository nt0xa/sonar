// Package batcher batches items by key with a growing window.
package batcher

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nt0xa/sonar/pkg/workerpool"
)

// Batcher groups items by key. The first passThrough items for a key are
// emitted immediately, subsequent items are buffered and emitted as a batch
// when the key's window expires. Every window that emits a batch doubles the next
// one (up to maxWindow); a window with nothing buffered resets the key.
// Batches are passed to the handler on a pool of worker goroutines.
type Batcher[T any] struct {
	keyFn KeyFn[T]
	opts  options

	mu      sync.Mutex
	entries map[string]*entry[T]
	pool    *workerpool.Pool[[]T]
	stopped bool

	dropped atomic.Int64
}

type KeyFn[T any] = func(T) string

type entry[T any] struct {
	buf    []T
	window time.Duration
	count  int
	timer  *time.Timer
}

// New creates a Batcher, panics on invalid arguments. The handler runs on
// worker goroutines and must not call Stop.
func New[T any](keyFn KeyFn[T], handler func([]T), opts ...Option) *Batcher[T] {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	if keyFn == nil {
		panic("batcher: keyFn must not be nil")
	}

	if handler == nil {
		panic("batcher: handler must not be nil")
	}

	if options.window <= 0 {
		panic("batcher: window must be > 0")
	}

	if options.maxBatch < 0 {
		panic("batcher: maxBatch must be >= 0")
	}

	if options.bufferSize < 0 {
		panic("batcher: bufferSize must be >= 0")
	}

	if options.passThrough < 1 {
		panic("batcher: passThrough must be >= 1")
	}

	if options.workers < 1 {
		panic("batcher: workers must be >= 1")
	}

	options.maxWindow = max(options.maxWindow, options.window)

	return &Batcher[T]{
		keyFn:   keyFn,
		opts:    options,
		entries: make(map[string]*entry[T]),
		pool: workerpool.New(options.workers, options.bufferSize,
			func(_ context.Context, batch []T) { handler(batch) },
			options.poolOpts...),
	}
}

// Push adds an item, never blocks; overflow is counted in Dropped.
func (b *Batcher[T]) Push(item T) {
	key := b.keyFn(item)

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.stopped {
		return
	}

	e, ok := b.entries[key]
	if !ok {
		b.entries[key] = &entry[T]{
			window: b.opts.window,
			count:  1,
			timer:  time.AfterFunc(b.opts.window, func() { b.tick(key) }),
		}
		b.emit([]T{item})
		return
	}

	if e.count < b.opts.passThrough {
		b.emit([]T{item})
		e.count++
		return
	}

	if len(e.buf) >= b.opts.maxBatch {
		b.dropped.Add(1)
		return
	}

	e.buf = append(e.buf, item)
}

func (b *Batcher[T]) tick(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.stopped {
		return
	}

	e := b.entries[key]

	if len(e.buf) == 0 {
		delete(b.entries, key)
		return
	}

	b.emit(e.buf)
	e.buf = nil
	e.window = min(e.window*2, b.opts.maxWindow)
	e.timer = time.AfterFunc(e.window, func() { b.tick(key) })
}

// emit must be called with b.mu held and b.stopped false, otherwise it may send to the stopped pool.
func (b *Batcher[T]) emit(batch []T) {
	if !b.pool.TryProcess(context.Background(), batch) {
		b.dropped.Add(int64(len(batch)))
	}
}

// Dropped returns the number of items dropped because a key's batch was
// full or the handler queue was full.
func (b *Batcher[T]) Dropped() int64 {
	return b.dropped.Load()
}

// Stop emits all pending batches and waits for the handlers to finish. If ctx
// is done first, its error is returned and handlers may still be running;
// calling Stop again waits again. Items pushed after Stop are ignored. Stop
// must not be called from the handler.
func (b *Batcher[T]) Stop(ctx context.Context) error {
	b.mu.Lock()

	if !b.stopped {
		b.stopped = true

		for _, e := range b.entries {
			e.timer.Stop()
			if len(e.buf) != 0 {
				b.emit(e.buf)
			}
		}
		clear(b.entries)
	}

	b.mu.Unlock()

	return b.pool.Stop(ctx)
}
