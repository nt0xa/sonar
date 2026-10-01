package coalescer_test

import (
	"iter"
	"strconv"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/pkg/coalescer"
)

func TestCoalescerSmoke(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := coalescer.New[int](
			func(i int) string { return "key" },
			coalescer.Window(time.Second),
			coalescer.MaxWindow(4*time.Second),
		)

		for i := range 5 {
			c.Push(i)
		}

		next, stop := iter.Pull(c.Next(t.Context()))
		defer stop()

		v, ok := next()
		require.True(t, ok)
		assert.Equal(t, []int{0}, v)

		v, ok = next()
		require.True(t, ok)
		assert.Equal(t, []int{1, 2, 3, 4}, v)
	})
}

func TestCoalescerBackoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := coalescer.New[int](
			func(i int) string { return "key" },
			coalescer.Window(time.Second),
			coalescer.MaxWindow(4*time.Second),
			coalescer.PassThrough(3),
		)

		start := time.Now()

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

		type emission struct {
			at time.Duration
			n  int
		}

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

		next, stop := iter.Pull(c.Next(t.Context()))
		defer stop()

		var got []emission
		for range want {
			v, ok := next()
			require.True(t, ok)
			got = append(got, emission{time.Since(start), len(v)})
		}

		assert.Equal(t, want, got)
	})
}

func TestCoalescerStop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := coalescer.New[int](
			func(i int) string { return "key" },
			coalescer.Window(time.Second),
			coalescer.MaxWindow(4*time.Second),
		)

		for i := range 3 {
			c.Push(i)
		}

		c.Stop()
		c.Stop()
		c.Push(3)

		// Pending timer fires after Stop.
		time.Sleep(10 * time.Second)

		var got [][]int
		for v := range c.Next(t.Context()) {
			got = append(got, v)
		}

		assert.Equal(t, [][]int{{0}, {1, 2}}, got)
	})
}

func TestCoalescerKeys(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := coalescer.New[int](
			func(i int) string { return strconv.Itoa(i % 2) },
			coalescer.Window(time.Second),
		)

		for i := range 6 {
			c.Push(i)
		}

		time.Sleep(10 * time.Second)
		c.Stop()

		var got [][]int
		for v := range c.Next(t.Context()) {
			got = append(got, v)
		}

		assert.ElementsMatch(t, [][]int{{0}, {1}, {2, 4}, {3, 5}}, got)
	})
}

func TestCoalescerSingleItem(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := coalescer.New[int](
			func(i int) string { return "key" },
			coalescer.Window(time.Second),
		)

		c.Push(0)

		// No empty batch after the window expires.
		time.Sleep(10 * time.Second)
		c.Stop()

		var got [][]int
		for v := range c.Next(t.Context()) {
			got = append(got, v)
		}

		assert.Equal(t, [][]int{{0}}, got)
	})
}

func TestCoalescerMaxBatch(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := coalescer.New[int](
			func(i int) string { return "key" },
			coalescer.Window(time.Second),
			coalescer.MaxBatch(2),
		)

		for i := range 5 {
			c.Push(i)
		}

		time.Sleep(10 * time.Second)
		c.Stop()

		var got [][]int
		for v := range c.Next(t.Context()) {
			got = append(got, v)
		}

		assert.Equal(t, [][]int{{0}, {1, 2}}, got)
		assert.EqualValues(t, 2, c.Dropped())
	})
}

func TestCoalescerBufferFull(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := coalescer.New[int](
			func(i int) string { return strconv.Itoa(i) },
			coalescer.BufferSize(1),
		)

		c.Push(0)
		c.Push(1)
		c.Stop()

		var got [][]int
		for v := range c.Next(t.Context()) {
			got = append(got, v)
		}

		assert.Equal(t, [][]int{{0}}, got)
		assert.EqualValues(t, 1, c.Dropped())
	})
}

func TestCoalescerInvalidOptions(t *testing.T) {
	keyFn := func(i int) string { return "key" }

	assert.Panics(t, func() { coalescer.New[int](nil) })
	assert.Panics(t, func() { coalescer.New(keyFn, coalescer.Window(0)) })
	assert.Panics(t, func() { coalescer.New(keyFn, coalescer.MaxBatch(-1)) })
	assert.Panics(t, func() { coalescer.New(keyFn, coalescer.BufferSize(-1)) })
	assert.Panics(t, func() { coalescer.New(keyFn, coalescer.PassThrough(0)) })
}
