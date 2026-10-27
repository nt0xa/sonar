// Package workerpool provides a generic buffered worker pool: items submitted
// via Submit are fanned out across a fixed number of worker goroutines, each
// running the handler.
package workerpool

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/time/rate"
)

// ErrStopped is returned by Submit when the pool is stopped.
var ErrStopped = errors.New("workerpool: stopped")

type Pool[T any] struct {
	limiter *rate.Limiter

	tasks   chan T
	handler func(T)
	wg      sync.WaitGroup

	// mu is held for reading by submitters and for writing by Stop to close tasks.
	mu       sync.RWMutex
	stopping chan struct{}

	stopOnce sync.Once
	done     chan struct{}
}

// New creates a Pool, panics on invalid arguments.
func New[T any](
	handler func(T),
	opts ...Option,
) *Pool[T] {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	if handler == nil {
		panic("workerpool: handler must not be nil")
	}

	if options.workers <= 0 {
		panic("workerpool: workers must be > 0")
	}

	if options.capacity < 0 {
		panic("workerpool: capacity must be >= 0")
	}

	var limiter *rate.Limiter

	if options.rateLimit {
		if options.limit <= 0 {
			panic("workerpool: rate limit must be > 0")
		}

		if options.burst <= 0 {
			panic("workerpool: rate limit burst must be > 0")
		}

		limiter = rate.NewLimiter(options.limit, options.burst)
	}

	p := Pool[T]{
		limiter:  limiter,
		tasks:    make(chan T, options.capacity),
		handler:  handler,
		stopping: make(chan struct{}),
		done:     make(chan struct{}),
	}

	p.wg.Add(options.workers)
	for range options.workers {
		go p.worker()
	}

	return &p
}

func (p *Pool[T]) worker() {
	defer p.wg.Done()

	for value := range p.tasks {
		if p.limiter != nil {
			// Can't fail: ctx is never cancelled and limit/burst are validated in New.
			_ = p.limiter.Wait(context.Background())
		}
		p.handler(value)
	}
}

// Submit enqueues value, blocking while the buffer is full. It returns ErrStopped
// if the pool is stopped.
func (p *Pool[T]) Submit(value T) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.isStopping() {
		return ErrStopped
	}

	select {
	case p.tasks <- value:
		return nil
	case <-p.stopping:
		return ErrStopped
	}
}

// TrySubmit is like Submit but doesn't block: it returns false if the buffer is
// full or the pool is stopped.
func (p *Pool[T]) TrySubmit(value T) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.isStopping() {
		return false
	}

	select {
	case p.tasks <- value:
		return true
	default:
		return false
	}
}

// isStopping must be called with p.mu held; while it is held and this returns
// false, tasks stays open.
func (p *Pool[T]) isStopping() bool {
	select {
	case <-p.stopping:
		return true
	default:
		return false
	}
}

// Stop rejects new submissions, handles buffered items and waits for the
// workers to finish. If ctx is done first, its error is returned and workers
// may still be running; calling Stop again waits again.
func (p *Pool[T]) Stop(ctx context.Context) error {
	p.stopOnce.Do(func() {
		// Unblock pending submitters before waiting for them to release mu.
		close(p.stopping)

		go func() {
			p.mu.Lock()
			close(p.tasks)
			p.mu.Unlock()

			p.wg.Wait()
			close(p.done)
		}()
	})

	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
