package coalescer_test

import (
	"context"
	"fmt"
	"time"

	"github.com/nt0xa/sonar/pkg/coalescer"
	"github.com/nt0xa/sonar/pkg/workerpool"
)

// Batches go straight into a worker pool; the pool's queue is the only buffer
// and a full queue counts as dropped in the coalescer.
func Example_workerpool() {
	pool := workerpool.New(1, 10, func(_ context.Context, batch []coalescer.Item[string]) {
		fmt.Println(len(batch))
	})

	c := coalescer.New(
		func(s string) string { return s },
		pool.TryProcess,
		coalescer.Window(time.Second),
	)

	for range 3 {
		c.Push(context.Background(), "key")
	}

	c.Stop() // flushes pending batches into the pool
	_ = pool.Stop(context.Background())

	// Output:
	// 1
	// 2
}
