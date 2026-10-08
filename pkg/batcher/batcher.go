// Package batcher batches items by key with a growing window.
package batcher

import (
	"sync"
	"time"
)

// Batcher groups items by key. The first passThrough items for a key are
// emitted immediately, subsequent items are buffered and emitted as a batch
// when the key's window expires. Every window that emits a batch doubles the next
// one (up to maxWindow); a window with nothing buffered resets the key.
// Batches are delivered on Batches.
type Batcher[T any, K comparable] struct {
	keyFn func(T) K
	opts  options

	mu      sync.Mutex
	entries map[K]*entry[T]
	out     chan []T
	closed  bool
}

type entry[T any] struct {
	buf    []T
	window time.Duration
	passed int
	timer  *time.Timer
}

// New creates a Batcher, panics on invalid arguments.
func New[T any, K comparable](
	keyFn func(T) K,
	opts ...Option,
) *Batcher[T, K] {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	if keyFn == nil {
		panic("batcher: keyFn must not be nil")
	}

	if options.window <= 0 {
		panic("batcher: window must be > 0")
	}

	if options.maxBatch < 0 {
		panic("batcher: maxBatch must be >= 0")
	}

	if options.outputCapacity < 0 {
		panic("batcher: outputCapacity must be >= 0")
	}

	if options.passThrough < 1 {
		panic("batcher: passThrough must be >= 1")
	}

	options.maxWindow = max(options.maxWindow, options.window)

	return &Batcher[T, K]{
		keyFn:   keyFn,
		opts:    options,
		entries: make(map[K]*entry[T]),
		out:     make(chan []T, options.outputCapacity),
	}
}

// Add adds an item and never blocks. It returns false if the item was dropped
// because the key's batch or the output channel was full, or b was closed.
func (b *Batcher[T, K]) Add(item T) bool {
	key := b.keyFn(item)

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return false
	}

	e, ok := b.entries[key]
	if !ok {
		b.entries[key] = &entry[T]{
			window: b.opts.window,
			passed: 1,
			timer:  time.AfterFunc(b.opts.window, func() { b.flushKey(key) }),
		}
		return b.tryEmit([]T{item})
	}

	if e.passed < b.opts.passThrough {
		e.passed++
		return b.tryEmit([]T{item})
	}

	if len(e.buf) >= b.opts.maxBatch {
		return false
	}

	e.buf = append(e.buf, item)
	return true
}

func (b *Batcher[T, K]) flushKey(key K) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	e, ok := b.entries[key]
	if !ok {
		return
	}

	if len(e.buf) == 0 {
		delete(b.entries, key)
		return
	}

	b.tryEmit(e.buf)
	e.buf = nil
	e.window = min(e.window*2, b.opts.maxWindow)
	e.timer = time.AfterFunc(e.window, func() { b.flushKey(key) })
}

// tryEmit must be called with b.mu held and b.closed false, otherwise it may send on the closed channel.
// It drops the batch and returns false if the channel is full.
func (b *Batcher[T, K]) tryEmit(batch []T) bool {
	select {
	case b.out <- batch:
		return true
	default:
		return false
	}
}

// Batches returns the channel batches are delivered on. It is closed by Close.
func (b *Batcher[T, K]) Batches() <-chan []T {
	return b.out
}

// Close emits all pending batches and closes the Batches channel.
// Items added after Close are dropped. Close is idempotent.
func (b *Batcher[T, K]) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.closed = true

	for _, e := range b.entries {
		e.timer.Stop()
		if len(e.buf) != 0 {
			b.tryEmit(e.buf)
		}
	}
	clear(b.entries)

	close(b.out)
}
