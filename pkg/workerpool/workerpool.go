// Package workerpool provides a generic buffered worker pool: items submitted
// via Process are fanned out across a fixed number of worker goroutines, each
// running the configured Handler.
package workerpool

import (
	"context"
	"sync"
)

type task[T any] struct {
	ctx   context.Context
	value T
}

type Processor[T any] struct {
	tasks     chan task[T]
	handler   func(context.Context, T)
	workersWg sync.WaitGroup

	stopOnce sync.Once
	done     chan struct{}
}

func NewProcessor[T any](
	workers int,
	capacity int,
	handler func(context.Context, T),
) *Processor[T] {
	if workers <= 0 {
		panic("workerpool: workers must be > 0")
	}

	if capacity < 0 {
		panic("workerpool: capacity must be >= 0")
	}

	if handler == nil {
		panic("workerpool: handler must not be nil")
	}

	p := Processor[T]{
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
func (p *Processor[T]) worker() {
	defer p.workersWg.Done()

	for task := range p.tasks {
		p.handler(task.ctx, task.value)
	}
}

// Process enqueues value for processing. The caller's ctx is preserved for its
// trace span and values but stripped of cancellation/deadline, so async
// processing isn't aborted when the originating interaction's ctx ends.
func (p *Processor[T]) Process(ctx context.Context, value T) {
	p.tasks <- task[T]{
		ctx:   context.WithoutCancel(ctx),
		value: value,
	}
}

func (p *Processor[T]) Stop(ctx context.Context) error {
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
