// Package workerpool provides a generic buffered worker pool: items submitted
// via Process are fanned out across a fixed number of worker goroutines, each
// running the configured Handler.
package workerpool

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

type task[T any] struct {
	ctx   context.Context
	value T
}

type Pool[T any] struct {
	limiter *rate.Limiter

	tasks     chan task[T]
	handler   func(context.Context, T)
	workersWg sync.WaitGroup

	stopOnce sync.Once
	done     chan struct{}
}

func New[T any](
	workers int,
	capacity int,
	handler func(context.Context, T),
	opts ...Option,
) *Pool[T] {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	if workers <= 0 {
		panic("workerpool: workers must be > 0")
	}

	if capacity < 0 {
		panic("workerpool: capacity must be >= 0")
	}

	if handler == nil {
		panic("workerpool: handler must not be nil")
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
		limiter: limiter,
		tasks:   make(chan task[T], capacity),
		handler: handler,
		done:    make(chan struct{}),
	}

	p.workersWg.Add(workers)
	for range workers {
		go p.worker()
	}

	return &p

}
func (p *Pool[T]) worker() {
	defer p.workersWg.Done()

	for task := range p.tasks {
		if p.limiter != nil {
			// Can't fail: ctx is never cancelled and limit/burst are validated in NewProcessor.
			_ = p.limiter.Wait(task.ctx)
		}
		p.handler(task.ctx, task.value)
	}
}

// Process enqueues value for processing. The caller's ctx is preserved for its
// trace span and values but stripped of cancellation/deadline, so async
// processing isn't aborted when the originating interaction's ctx ends.
func (p *Pool[T]) Process(ctx context.Context, value T) {
	p.tasks <- task[T]{
		ctx:   context.WithoutCancel(ctx),
		value: value,
	}
}

// TryProcess is like Process but returns false instead of blocking when the buffer is full.
// Like Process, it must not be called concurrently with or after Stop: sending
// on the closed queue panics.
func (p *Pool[T]) TryProcess(ctx context.Context, value T) bool {
	select {
	case p.tasks <- task[T]{ctx: context.WithoutCancel(ctx), value: value}:
		return true
	default:
		return false
	}
}

func (p *Pool[T]) Stop(ctx context.Context) error {
	p.stopOnce.Do(func() {
		close(p.tasks)

		go func() {
			p.workersWg.Wait()
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
