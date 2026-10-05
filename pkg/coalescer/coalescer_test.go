package coalescer_test

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/pkg/coalescer"
)

func values[T any](batch []coalescer.Item[T]) []T {
	vs := make([]T, len(batch))
	for i, it := range batch {
		vs[i] = it.Value
	}
	return vs
}

// collector is a handler that records the values of every batch.
type collector struct {
	mu  sync.Mutex
	got [][]int
}

func (c *collector) handle(batch []coalescer.Item[int]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.got = append(c.got, values(batch))
}

func (c *collector) batches() [][]int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.got
}

func Test_Smoke(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var col collector
		c := coalescer.New(
			func(i int) string { return "key" },
			col.handle,
			coalescer.Window(time.Second),
			coalescer.MaxWindow(4*time.Second),
		)

		for i := range 5 {
			c.Push(t.Context(), i)
		}

		time.Sleep(time.Second)
		require.NoError(t, c.Stop(t.Context()))

		assert.Equal(t, [][]int{{0}, {1, 2, 3, 4}}, col.batches())
	})
}

func Test_Backoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		type emission struct {
			at time.Duration
			n  int
		}

		var (
			mu    sync.Mutex
			got   []emission
			start = time.Now()
		)

		// With a single worker and the fake clock, handling time equals emission time.
		c := coalescer.New(
			func(i int) string { return "key" },
			func(batch []coalescer.Item[int]) {
				mu.Lock()
				defer mu.Unlock()
				got = append(got, emission{time.Since(start), len(batch)})
			},
			coalescer.Window(time.Second),
			coalescer.MaxWindow(4*time.Second),
			coalescer.PassThrough(3),
		)

		go func() {
			// Continuous traffic until 10s, offset to avoid window boundaries.
			c.Push(t.Context(), 0)
			time.Sleep(100 * time.Millisecond)
			for i := 1; time.Since(start) < 10*time.Second; i++ {
				c.Push(t.Context(), i)
				time.Sleep(250 * time.Millisecond)
			}

			// Quiet period resets the key: pass-through and windows start over.
			time.Sleep(20*time.Second - time.Since(start))
			for i := range 4 {
				c.Push(t.Context(), 100+i)
				time.Sleep(100 * time.Millisecond)
			}
		}()

		want := []emission{
			{0, 1},                                     // pass-through
			{100 * time.Millisecond, 1},                // pass-through
			{350 * time.Millisecond, 1},                // pass-through
			{1 * time.Second, 2},                       // window 1s
			{3 * time.Second, 8},                       // window 2s
			{7 * time.Second, 16},                      // window 4s
			{11 * time.Second, 12},                     // window capped at 4s
			{20 * time.Second, 1},                      // reset: pass-through
			{20*time.Second + 100*time.Millisecond, 1}, // pass-through
			{20*time.Second + 200*time.Millisecond, 1}, // pass-through
			{21 * time.Second, 1},                      // window 1s again
		}

		time.Sleep(30 * time.Second)
		require.NoError(t, c.Stop(t.Context()))

		assert.Equal(t, want, got)
	})
}

func Test_Stop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var col collector
		c := coalescer.New(
			func(i int) string { return "key" },
			col.handle,
			coalescer.Window(time.Second),
			coalescer.MaxWindow(4*time.Second),
		)

		for i := range 3 {
			c.Push(t.Context(), i)
		}

		require.NoError(t, c.Stop(t.Context()))
		require.NoError(t, c.Stop(t.Context()))
		c.Push(t.Context(), 3)

		// The pending window would have fired here.
		time.Sleep(10 * time.Second)

		assert.Equal(t, [][]int{{0}, {1, 2}}, col.batches())
	})
}

func Test_StopDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		c := coalescer.New(
			func(i int) string { return "key" },
			func([]coalescer.Item[int]) { <-release },
		)

		c.Push(t.Context(), 0)
		synctest.Wait() // the handler is blocked

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		assert.ErrorIs(t, c.Stop(ctx), context.DeadlineExceeded)

		// Shutdown keeps going; a later Stop waits for it to finish.
		close(release)
		assert.NoError(t, c.Stop(t.Context()))
	})
}

func Test_Keys(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var col collector
		c := coalescer.New(
			func(i int) string { return strconv.Itoa(i % 2) },
			col.handle,
			coalescer.Window(time.Second),
		)

		for i := range 6 {
			c.Push(t.Context(), i)
		}

		time.Sleep(10 * time.Second)
		require.NoError(t, c.Stop(t.Context()))

		assert.ElementsMatch(t, [][]int{{0}, {1}, {2, 4}, {3, 5}}, col.batches())
	})
}

func Test_SingleItem(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var col collector
		c := coalescer.New(
			func(i int) string { return "key" },
			col.handle,
			coalescer.Window(time.Second),
		)

		c.Push(t.Context(), 0)

		// No empty batch after the window expires.
		time.Sleep(10 * time.Second)
		require.NoError(t, c.Stop(t.Context()))

		assert.Equal(t, [][]int{{0}}, col.batches())
	})
}

func Test_MaxBatch(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var col collector
		c := coalescer.New(
			func(i int) string { return "key" },
			col.handle,
			coalescer.Window(time.Second),
			coalescer.MaxBatch(2),
		)

		for i := range 5 {
			c.Push(t.Context(), i)
		}

		time.Sleep(10 * time.Second)
		require.NoError(t, c.Stop(t.Context()))

		assert.Equal(t, [][]int{{0}, {1, 2}}, col.batches())
		assert.EqualValues(t, 2, c.Dropped())
	})
}

func Test_BufferFull(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var col collector
		release := make(chan struct{})
		c := coalescer.New(
			func(i int) string { return "key" },
			func(batch []coalescer.Item[int]) {
				<-release
				col.handle(batch)
			},
			coalescer.BufferSize(1),
			coalescer.PassThrough(3),
		)

		c.Push(t.Context(), 0)
		synctest.Wait() // the worker holds 0

		c.Push(t.Context(), 1) // queued
		c.Push(t.Context(), 2) // dropped: queue full

		close(release)
		require.NoError(t, c.Stop(t.Context()))

		assert.Equal(t, [][]int{{0}, {1}}, col.batches())
		assert.EqualValues(t, 1, c.Dropped())
	})
}

func Test_PushStripsCancellationButKeepsValues(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		type key struct{}

		var (
			mu  sync.Mutex
			got []coalescer.Item[int]
		)

		c := coalescer.New(
			func(i int) string { return "key" },
			func(batch []coalescer.Item[int]) {
				mu.Lock()
				defer mu.Unlock()
				got = append(got, batch...)
			},
			coalescer.Window(time.Second),
		)

		for i := range 3 {
			ctx, cancel := context.WithCancel(context.WithValue(t.Context(), key{}, i))
			c.Push(ctx, i)
			cancel() // the originating interaction ends before the item is handled
		}

		time.Sleep(10 * time.Second)
		require.NoError(t, c.Stop(t.Context()))

		require.Len(t, got, 3)
		for _, it := range got {
			assert.Equal(t, it.Value, it.Ctx.Value(key{}))
			assert.NoError(t, it.Ctx.Err())
		}
	})
}

func Test_InvalidOptions(t *testing.T) {
	keyFn := func(i int) string { return "key" }
	handler := func([]coalescer.Item[int]) {}

	assert.Panics(t, func() { coalescer.New(nil, handler) })
	assert.Panics(t, func() { coalescer.New(keyFn, nil) })
	assert.Panics(t, func() { coalescer.New(keyFn, handler, coalescer.Window(0)) })
	assert.Panics(t, func() { coalescer.New(keyFn, handler, coalescer.MaxBatch(-1)) })
	assert.Panics(t, func() { coalescer.New(keyFn, handler, coalescer.BufferSize(-1)) })
	assert.Panics(t, func() { coalescer.New(keyFn, handler, coalescer.PassThrough(0)) })
	assert.Panics(t, func() { coalescer.New(keyFn, handler, coalescer.Workers(0)) })
}
