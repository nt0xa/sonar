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

func TestProcessor_Panics(t *testing.T) {
	require.Panics(t, func() {
		_ = workerpool.New(0, 1, func(context.Context, int) {})
	})

	require.Panics(t, func() {
		_ = workerpool.New(1, -1, func(context.Context, int) {})
	})

	require.Panics(t, func() {
		_ = workerpool.New[int](1, 1, nil)
	})

	require.Panics(t, func() {
		_ = workerpool.New(1, 1, func(context.Context, int) {}, workerpool.RateLimit(0, 1))
	})

	require.Panics(t, func() {
		_ = workerpool.New(1, 1, func(context.Context, int) {}, workerpool.RateLimit(1, 0))
	})
}

func TestProcessor_HandlesEverySubmittedItem(t *testing.T) {
	var (
		mu   sync.Mutex
		seen []int
	)

	p := workerpool.New(4, 16, func(_ context.Context, v int) {
		mu.Lock()
		defer mu.Unlock()
		seen = append(seen, v)
	})

	for i := range 100 {
		p.Process(t.Context(), i)
	}

	require.NoError(t, p.Stop(t.Context()))

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, seen, 100)
}

func TestProcessor_StopDrainsBufferedItems(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var handled atomic.Int64

		// One slow worker and a deep buffer: everything is still queued when Stop
		// is called, so a Stop that didn't drain would lose it.
		p := workerpool.New(1, 64, func(_ context.Context, _ int) {
			time.Sleep(time.Millisecond)
			handled.Add(1)
		})

		for i := range 50 {
			p.Process(t.Context(), i)
		}

		require.NoError(t, p.Stop(t.Context()))
		assert.EqualValues(t, 50, handled.Load())
	})
}

func TestProcessor_StopGivesUpWhenContextIsDone(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})

		p := workerpool.New(1, 4, func(_ context.Context, _ int) {
			<-release
		})

		p.Process(t.Context(), 1)

		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		defer cancel()

		assert.ErrorIs(t, p.Stop(ctx), context.DeadlineExceeded)

		close(release)
	})
}

func TestProcessor_ProcessStripsCancellationButKeepsValues(t *testing.T) {
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

	p := workerpool.New(1, 1, handler)

	ctx, cancel := context.WithCancel(context.WithValue(t.Context(), key{}, "kept"))
	p.Process(ctx, 1)
	cancel() // the originating interaction ends before the item is handled

	<-done
	require.NoError(t, p.Stop(t.Context()))

	assert.Equal(t, "kept", gotVal)
	assert.NoError(t, gotErr)
}

func TestProcessor_RateLimitIsSharedAcrossWorkers(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const (
			items    = 5
			interval = 20 * time.Millisecond
		)

		var handled atomic.Int64

		// Burst 1 and several workers: the limiter must pace all of them together,
		// not each one separately.
		p := workerpool.New(4, items, func(_ context.Context, _ int) {
			handled.Add(1)
		}, workerpool.RateLimit(rate.Every(interval), 1))

		start := time.Now()

		for i := range items {
			p.Process(t.Context(), i)
		}

		require.NoError(t, p.Stop(t.Context()))

		assert.EqualValues(t, items, handled.Load())
		assert.Equal(t, (items-1)*interval, time.Since(start))
	})
}
