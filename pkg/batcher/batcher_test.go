package batcher_test

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/nt0xa/sonar/pkg/batcher"
)

type collector struct {
	mu      sync.Mutex
	batches [][]int
}

func (c *collector) handle(batch []int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.batches = append(c.batches, batch)
}

func Test_Smoke(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var col collector

		c := batcher.New(
			func(i int) string { return "key" },
			col.handle,
			batcher.Window(time.Second),
			batcher.MaxWindow(4*time.Second),
		)

		for i := range 5 {
			c.Push(i)
		}

		time.Sleep(10 * time.Second)
		require.NoError(t, c.Stop(t.Context()))

		assert.Equal(t, [][]int{{0}, {1, 2, 3, 4}}, col.batches)
	})
}

func Test_Backoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		type emission struct {
			at time.Duration
			n  int
		}

		var (
			start = time.Now()
			mu    sync.Mutex
			got   []emission
		)

		c := batcher.New(
			func(i int) string { return "key" },
			func(batch []int) {
				mu.Lock()
				defer mu.Unlock()
				got = append(got, emission{time.Since(start), len(batch)})
			},
			batcher.Window(time.Second),
			batcher.MaxWindow(4*time.Second),
			batcher.PassThrough(3),
		)

		go func() {
			// Continuous traffic until 10s, offset to avoid window boundaries.
			c.Push(0)
			time.Sleep(100 * time.Millisecond)
			for i := 1; time.Since(start) < 10*time.Second; i++ {
				c.Push(i)
				time.Sleep(250 * time.Millisecond)
			}

			// Quiet period resets the key: pass-through and windows start over.
			time.Sleep(20*time.Second - time.Since(start))
			for i := range 4 {
				c.Push(100 + i)
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

		c := batcher.New(
			func(i int) string { return "key" },
			col.handle,
			batcher.Window(time.Second),
			batcher.MaxWindow(4*time.Second),
		)

		for i := range 3 {
			c.Push(i)
		}

		require.NoError(t, c.Stop(t.Context()))
		require.NoError(t, c.Stop(t.Context()))
		c.Push(3)

		// Nothing is emitted after Stop.
		time.Sleep(10 * time.Second)

		assert.Equal(t, [][]int{{0}, {1, 2}}, col.batches)
	})
}

func Test_Keys(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var col collector

		c := batcher.New(
			func(i int) string { return strconv.Itoa(i % 2) },
			col.handle,
			batcher.Window(time.Second),
		)

		for i := range 6 {
			c.Push(i)
		}

		time.Sleep(10 * time.Second)
		require.NoError(t, c.Stop(t.Context()))

		assert.ElementsMatch(t, [][]int{{0}, {1}, {2, 4}, {3, 5}}, col.batches)
	})
}

func Test_SingleItem(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var col collector

		c := batcher.New(
			func(i int) string { return "key" },
			col.handle,
			batcher.Window(time.Second),
		)

		c.Push(0)

		// No empty batch after the window expires.
		time.Sleep(10 * time.Second)
		require.NoError(t, c.Stop(t.Context()))

		assert.Equal(t, [][]int{{0}}, col.batches)
	})
}

func Test_MaxBatch(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var col collector

		c := batcher.New(
			func(i int) string { return "key" },
			col.handle,
			batcher.Window(time.Second),
			batcher.MaxBatch(2),
		)

		for i := range 5 {
			c.Push(i)
		}

		time.Sleep(10 * time.Second)
		require.NoError(t, c.Stop(t.Context()))

		assert.Equal(t, [][]int{{0}, {1, 2}}, col.batches)
		assert.EqualValues(t, 2, c.Dropped())
	})
}

func Test_BufferFull(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			col     collector
			release = make(chan struct{})
		)

		c := batcher.New(
			func(i int) string { return "key" },
			func(batch []int) {
				<-release
				col.handle(batch)
			},
			batcher.PassThrough(3),
			batcher.BufferSize(1),
		)

		c.Push(0)
		synctest.Wait() // the worker takes 0 and blocks

		c.Push(1) // queued
		c.Push(2) // queue full

		close(release)
		require.NoError(t, c.Stop(t.Context()))

		assert.Equal(t, [][]int{{0}, {1}}, col.batches)
		assert.EqualValues(t, 1, c.Dropped())
	})
}

func Test_StopTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			col     collector
			release = make(chan struct{})
		)

		c := batcher.New(
			func(i int) string { return "key" },
			func(batch []int) {
				<-release
				col.handle(batch)
			},
		)

		c.Push(0)
		synctest.Wait() // the worker takes 0 and blocks

		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		defer cancel()

		assert.ErrorIs(t, c.Stop(ctx), context.DeadlineExceeded)

		close(release)
		require.NoError(t, c.Stop(t.Context()))

		assert.Equal(t, [][]int{{0}}, col.batches)
	})
}

func Test_RateLimit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			start = time.Now()
			mu    sync.Mutex
			got   []time.Duration
		)

		c := batcher.New(
			func(i int) string { return strconv.Itoa(i) },
			func([]int) {
				mu.Lock()
				defer mu.Unlock()
				got = append(got, time.Since(start))
			},
			batcher.Workers(3),
			batcher.RateLimit(rate.Every(time.Second), 1),
		)

		// Distinct keys pass through at once, the limiter spaces out the handler.
		for i := range 3 {
			c.Push(i)
		}

		require.NoError(t, c.Stop(t.Context()))

		assert.Equal(t, []time.Duration{0, time.Second, 2 * time.Second}, got)
	})
}

func Test_InvalidOptions(t *testing.T) {
	keyFn := func(i int) string { return "key" }
	handler := func([]int) {}

	assert.Panics(t, func() { batcher.New(nil, handler) })
	assert.Panics(t, func() { batcher.New(keyFn, nil) })
	assert.Panics(t, func() { batcher.New(keyFn, handler, batcher.Window(0)) })
	assert.Panics(t, func() { batcher.New(keyFn, handler, batcher.MaxBatch(-1)) })
	assert.Panics(t, func() { batcher.New(keyFn, handler, batcher.BufferSize(-1)) })
	assert.Panics(t, func() { batcher.New(keyFn, handler, batcher.PassThrough(0)) })
	assert.Panics(t, func() { batcher.New(keyFn, handler, batcher.Workers(0)) })
	assert.Panics(t, func() { batcher.New(keyFn, handler, batcher.RateLimit(0, 1)) })
}
