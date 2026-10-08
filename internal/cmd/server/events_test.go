package server

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/modules"
	"github.com/nt0xa/sonar/pkg/batcher"
	"github.com/nt0xa/sonar/pkg/telemetry"
)

type fakeNotifier struct {
	mu      sync.Mutex
	single  int
	batches []int
}

func (f *fakeNotifier) Name() string { return "fake" }

func (f *fakeNotifier) RateLimit() (rate.Limit, int) { return rate.Inf, 1 }

func (f *fakeNotifier) Notify(context.Context, *modules.Notification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.single++
	return nil
}

func (f *fakeNotifier) NotifyBatch(_ context.Context, ns []*modules.Notification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.batches = append(f.batches, len(ns))
	return nil
}

func Test_NotifierBatching(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		log := slog.New(slog.DiscardHandler)
		h := NewEventsHandler(nil, nil, log, 1, 1)

		var f fakeNotifier
		const passThrough = 5
		h.AddNotifier("fake", NewNotifier(&f, log, telemetry.NewNoop(), 1, batcher.WithPassThrough(passThrough)))

		n := &modules.Notification{
			User:    &database.User{},
			Payload: &database.Payload{ID: 1},
			Event:   &database.Event{Protocol: "dns"},
		}

		for range passThrough + 3 {
			h.notifiers["fake"].Add(t.Context(), n)
		}

		time.Sleep(10 * time.Second)
		h.notifiers["fake"].Close()
		synctest.Wait() // the pool handles the remaining batches and stops
		require.NoError(t, h.proc.Stop(t.Context()))

		assert.Equal(t, passThrough, f.single)
		assert.Equal(t, []int{3}, f.batches)
	})
}
