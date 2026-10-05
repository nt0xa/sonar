package coalescer_test

import (
	"strconv"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nt0xa/sonar/pkg/coalescer"
)

type emission struct {
	at    time.Duration
	batch []int
}

// recorder is an EmitFn that records every batch it accepts.
type recorder struct {
	start  time.Time
	reject bool

	mu  sync.Mutex
	got []emission
}

func newRecorder() *recorder {
	return &recorder{start: time.Now()}
}

func (r *recorder) emit(batch []int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.reject {
		return false
	}

	r.got = append(r.got, emission{time.Since(r.start), batch})
	return true
}

func (r *recorder) setReject(v bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reject = v
}

func (r *recorder) emissions() []emission {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]emission(nil), r.got...)
}

func (r *recorder) values() [][]int {
	var vs [][]int
	for _, e := range r.emissions() {
		vs = append(vs, e.batch)
	}
	return vs
}

func Test_Smoke(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRecorder()
		c := coalescer.New(
			func(i int) string { return "key" },
			r.emit,
			coalescer.Window(time.Second),
			coalescer.MaxWindow(4*time.Second),
		)

		for i := range 5 {
			c.Push(i)
		}

		assert.Equal(t, [][]int{{0}}, r.values())

		time.Sleep(time.Second)
		synctest.Wait() // let the window timer finish
		assert.Equal(t, [][]int{{0}, {1, 2, 3, 4}}, r.values())
	})
}

func Test_Backoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRecorder()
		c := coalescer.New(
			func(i int) string { return "key" },
			r.emit,
			coalescer.Window(time.Second),
			coalescer.MaxWindow(4*time.Second),
			coalescer.PassThrough(3),
		)

		start := r.start

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

		time.Sleep(time.Minute)

		type point struct {
			at time.Duration
			n  int
		}

		want := []point{
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

		var got []point
		for _, e := range r.emissions() {
			got = append(got, point{e.at, len(e.batch)})
		}

		assert.Equal(t, want, got)
	})
}

func Test_Stop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRecorder()
		c := coalescer.New(
			func(i int) string { return "key" },
			r.emit,
			coalescer.Window(time.Second),
			coalescer.MaxWindow(4*time.Second),
		)

		for i := range 3 {
			c.Push(i)
		}

		c.Stop()
		assert.Equal(t, [][]int{{0}, {1, 2}}, r.values())

		c.Stop()
		c.Push(3)

		// Pending timer fires after Stop.
		time.Sleep(10 * time.Second)

		assert.Equal(t, [][]int{{0}, {1, 2}}, r.values())
	})
}

func Test_Keys(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRecorder()
		c := coalescer.New(
			func(i int) string { return strconv.Itoa(i % 2) },
			r.emit,
			coalescer.Window(time.Second),
		)

		for i := range 6 {
			c.Push(i)
		}

		time.Sleep(10 * time.Second)

		assert.ElementsMatch(t, [][]int{{0}, {1}, {2, 4}, {3, 5}}, r.values())
	})
}

func Test_SingleItem(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRecorder()
		c := coalescer.New(
			func(i int) string { return "key" },
			r.emit,
			coalescer.Window(time.Second),
		)

		c.Push(0)

		// No empty batch after the window expires.
		time.Sleep(10 * time.Second)

		assert.Equal(t, [][]int{{0}}, r.values())
	})
}

func Test_MaxBatch(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRecorder()
		c := coalescer.New(
			func(i int) string { return "key" },
			r.emit,
			coalescer.Window(time.Second),
			coalescer.MaxBatch(2),
		)

		for i := range 5 {
			c.Push(i)
		}

		time.Sleep(10 * time.Second)

		assert.Equal(t, [][]int{{0}, {1, 2}}, r.values())
		assert.EqualValues(t, 2, c.Dropped())
	})
}

func Test_EmitRejected(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRecorder()
		c := coalescer.New(
			func(i int) string { return "key" },
			r.emit,
			coalescer.Window(time.Second),
		)

		c.Push(0)

		r.setReject(true)
		c.Push(1)
		c.Push(2)
		time.Sleep(10 * time.Second)

		assert.Equal(t, [][]int{{0}}, r.values())
		assert.EqualValues(t, 2, c.Dropped())
	})
}

func Test_InvalidOptions(t *testing.T) {
	keyFn := func(i int) string { return "key" }
	emitFn := func([]int) bool { return true }

	assert.Panics(t, func() { coalescer.New(nil, emitFn) })
	assert.Panics(t, func() { coalescer.New(keyFn, nil) })
	assert.Panics(t, func() { coalescer.New(keyFn, emitFn, coalescer.Window(0)) })
	assert.Panics(t, func() { coalescer.New(keyFn, emitFn, coalescer.MaxBatch(-1)) })
	assert.Panics(t, func() { coalescer.New(keyFn, emitFn, coalescer.PassThrough(0)) })
}
