package batcher_test

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nt0xa/sonar/pkg/batcher"
)

// collect drains b's batches; the returned func waits for b to be closed.
func collect[T any, K comparable](b *batcher.Batcher[T, K]) func() [][]T {
	var (
		got  [][]T
		done = make(chan struct{})
	)

	go func() {
		defer close(done)
		for batch := range b.Batches() {
			got = append(got, batch)
		}
	}()

	return func() [][]T {
		<-done
		return got
	}
}

func Test_Smoke(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := batcher.New(
			func(i int) string { return "key" },
			batcher.WithWindow(time.Second),
			batcher.WithMaxWindow(4*time.Second),
		)
		got := collect(c)

		for i := range 5 {
			c.Add(i)
		}

		time.Sleep(10 * time.Second)
		c.Close()

		assert.Equal(t, [][]int{{0}, {1, 2, 3, 4}}, got())
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
			got   []emission
			done  = make(chan struct{})
		)

		c := batcher.New(
			func(i int) string { return "key" },
			batcher.WithWindow(time.Second),
			batcher.WithMaxWindow(4*time.Second),
			batcher.WithPassThrough(3),
		)

		go func() {
			defer close(done)
			for batch := range c.Batches() {
				got = append(got, emission{time.Since(start), len(batch)})
			}
		}()

		go func() {
			// Continuous traffic until 10s, offset to avoid window boundaries.
			c.Add(0)
			time.Sleep(100 * time.Millisecond)
			for i := 1; time.Since(start) < 10*time.Second; i++ {
				c.Add(i)
				time.Sleep(250 * time.Millisecond)
			}

			// Quiet period resets the key: pass-through and windows start over.
			time.Sleep(20*time.Second - time.Since(start))
			for i := range 4 {
				c.Add(100 + i)
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
		c.Close()
		<-done

		assert.Equal(t, want, got)
	})
}

func Test_Close(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := batcher.New(
			func(i int) string { return "key" },
			batcher.WithWindow(time.Second),
			batcher.WithMaxWindow(4*time.Second),
		)
		got := collect(c)

		for i := range 3 {
			c.Add(i)
		}

		c.Close()
		c.Close()
		assert.False(t, c.Add(3))

		// Nothing is emitted after Close.
		time.Sleep(10 * time.Second)

		assert.Equal(t, [][]int{{0}, {1, 2}}, got())
	})
}

func Test_Keys(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := batcher.New(
			func(i int) int { return i % 2 },
			batcher.WithWindow(time.Second),
		)
		got := collect(c)

		for i := range 6 {
			c.Add(i)
		}

		time.Sleep(10 * time.Second)
		c.Close()

		assert.ElementsMatch(t, [][]int{{0}, {1}, {2, 4}, {3, 5}}, got())
	})
}

func Test_SingleItem(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := batcher.New(
			func(i int) string { return "key" },
			batcher.WithWindow(time.Second),
		)
		got := collect(c)

		c.Add(0)

		// No empty batch after the window expires.
		time.Sleep(10 * time.Second)
		c.Close()

		assert.Equal(t, [][]int{{0}}, got())
	})
}

func Test_MaxBatch(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := batcher.New(
			func(i int) string { return "key" },
			batcher.WithWindow(time.Second),
			batcher.WithMaxBatch(2),
		)
		got := collect(c)

		var accepted []bool
		for i := range 5 {
			accepted = append(accepted, c.Add(i))
		}

		time.Sleep(10 * time.Second)
		c.Close()

		assert.Equal(t, []bool{true, true, true, false, false}, accepted)
		assert.Equal(t, [][]int{{0}, {1, 2}}, got())
	})
}

func Test_BufferFull(t *testing.T) {
	c := batcher.New(
		func(i int) string { return "key" },
		batcher.WithPassThrough(3),
		batcher.WithOutputCapacity(1),
	)

	assert.True(t, c.Add(0))  // queued
	assert.False(t, c.Add(1)) // channel full

	c.Close()

	assert.Equal(t, [][]int{{0}}, collect(c)())
}

func Test_InvalidOptions(t *testing.T) {
	keyFn := func(i int) string { return "key" }

	assert.Panics(t, func() { batcher.New[int, string](nil) })
	assert.Panics(t, func() { batcher.New(keyFn, batcher.WithWindow(0)) })
	assert.Panics(t, func() { batcher.New(keyFn, batcher.WithMaxBatch(-1)) })
	assert.Panics(t, func() { batcher.New(keyFn, batcher.WithOutputCapacity(-1)) })
	assert.Panics(t, func() { batcher.New(keyFn, batcher.WithPassThrough(0)) })
}
