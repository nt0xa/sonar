package workerpool_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/nt0xa/sonar/pkg/workerpool"
)

func Test_Panics(t *testing.T) {
	procFn := func(context.Context, int) {}

	require.Panics(t, func() { workerpool.New(procFn, workerpool.WithWorkers(0)) })
	require.Panics(t, func() { workerpool.New(procFn, workerpool.WithCapacity(-1)) })
	require.Panics(t, func() { workerpool.New[int](nil) })
	require.Panics(t, func() { workerpool.New(procFn, workerpool.WithRateLimit(0, 1)) })
	require.Panics(t, func() { workerpool.New(procFn, workerpool.WithRateLimit(1, 0)) })
}

func Test_HandlesEverySubmittedItem(t *testing.T) {
	var (
		mu   sync.Mutex
		seen []int
	)

	p := workerpool.New(func(_ context.Context, v int) {
		mu.Lock()
		defer mu.Unlock()
		seen = append(seen, v)
	}, workerpool.WithWorkers(4), workerpool.WithCapacity(16))

	for i := range 100 {
		require.NoError(t, p.Submit(t.Context(), i))
	}

	require.NoError(t, p.Stop(t.Context()))

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, seen, 100)
}

func Test_StopDrainsBufferedItems(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var handled atomic.Int64

		// One slow worker and a deep buffer: everything is still queued when Stop
		// is called, so a Stop that didn't drain would lose it.
		p := workerpool.New(func(_ context.Context, _ int) {
			time.Sleep(time.Millisecond)
			handled.Add(1)
		}, workerpool.WithCapacity(64))

		for i := range 50 {
			require.NoError(t, p.Submit(t.Context(), i))
		}

		require.NoError(t, p.Stop(t.Context()))
		assert.EqualValues(t, 50, handled.Load())
	})
}

func Test_StopGivesUpWhenContextIsDone(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})

		p := workerpool.New(func(_ context.Context, _ int) {
			<-release
		}, workerpool.WithCapacity(4))

		require.NoError(t, p.Submit(t.Context(), 1))

		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		defer cancel()

		assert.ErrorIs(t, p.Stop(ctx), context.DeadlineExceeded)

		close(release)
	})
}

func Test_SubmitStripsCancellationButKeepsValues(t *testing.T) {
	type key struct{}

	var (
		done    = make(chan struct{})
		gotVal  any
		gotErr  error
		handler = func(ctx context.Context, _ int) {
			gotVal = ctx.Value(key{})
			gotErr = ctx.Err()
			close(done)
		}
	)

	p := workerpool.New(handler, workerpool.WithCapacity(1))

	ctx, cancel := context.WithCancel(context.WithValue(t.Context(), key{}, "kept"))
	require.NoError(t, p.Submit(ctx, 1))
	cancel() // the originating interaction ends before the item is handled

	<-done
	require.NoError(t, p.Stop(t.Context()))

	assert.Equal(t, "kept", gotVal)
	assert.NoError(t, gotErr)
}

func Test_SubmitGivesUpWhenContextIsDone(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})

		p := workerpool.New(func(_ context.Context, _ int) {
			<-release
		})

		require.NoError(t, p.Submit(t.Context(), 1))
		synctest.Wait() // the worker takes 1 and blocks

		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		defer cancel()

		assert.ErrorIs(t, p.Submit(ctx, 2), context.DeadlineExceeded)

		close(release)
		require.NoError(t, p.Stop(t.Context()))
	})
}

func Test_SubmitAfterStop(t *testing.T) {
	p := workerpool.New(func(context.Context, int) {}, workerpool.WithCapacity(1))

	require.NoError(t, p.Stop(t.Context()))

	assert.ErrorIs(t, p.Submit(t.Context(), 1), workerpool.ErrStopped)
	assert.False(t, p.TrySubmit(t.Context(), 1))
}

func Test_StopUnblocksPendingSubmit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			release = make(chan struct{})
			errCh   = make(chan error)
		)

		p := workerpool.New(func(_ context.Context, _ int) {
			<-release
		})

		require.NoError(t, p.Submit(t.Context(), 1))
		synctest.Wait() // the worker takes 1 and blocks

		go func() { errCh <- p.Submit(t.Context(), 2) }()
		synctest.Wait() // Submit blocks: no free worker, no buffer

		stopped := make(chan error)
		go func() { stopped <- p.Stop(t.Context()) }()

		assert.ErrorIs(t, <-errCh, workerpool.ErrStopped)

		close(release)
		require.NoError(t, <-stopped)
	})
}

func Test_RateLimitIsSharedAcrossWorkers(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const (
			items    = 5
			interval = 20 * time.Millisecond
		)

		var handled atomic.Int64

		// Burst 1 and several workers: the limiter must pace all of them together,
		// not each one separately.
		p := workerpool.New(func(_ context.Context, _ int) {
			handled.Add(1)
		},
			workerpool.WithWorkers(4),
			workerpool.WithCapacity(items),
			workerpool.WithRateLimit(rate.Every(interval), 1),
		)

		start := time.Now()

		for i := range items {
			require.NoError(t, p.Submit(t.Context(), i))
		}

		require.NoError(t, p.Stop(t.Context()))

		assert.EqualValues(t, items, handled.Load())
		assert.Equal(t, (items-1)*interval, time.Since(start))
	})
}

func Test_TrySubmitDropsWhenFull(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			release = make(chan struct{})
			handled []int
		)

		p := workerpool.New(func(_ context.Context, v int) {
			<-release
			handled = append(handled, v)
		}, workerpool.WithCapacity(1))

		assert.True(t, p.TrySubmit(t.Context(), 1))
		synctest.Wait() // the worker takes 1 and blocks

		assert.True(t, p.TrySubmit(t.Context(), 2))  // buffered
		assert.False(t, p.TrySubmit(t.Context(), 3)) // buffer full

		close(release)
		require.NoError(t, p.Stop(t.Context()))

		assert.Equal(t, []int{1, 2}, handled)
	})
}
