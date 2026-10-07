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
		h := NewEventsHandler(nil, nil, slog.New(slog.DiscardHandler), telemetry.NewNoop(), 1, 1)

		var f fakeNotifier
		h.AddNotifier("fake", &f)

		n := &modules.Notification{
			User:    &database.User{},
			Payload: &database.Payload{ID: 1},
			Event:   &database.Event{Protocol: "dns"},
		}

		for range notifyPassThrough + 3 {
			h.notifiers["fake"].Push(notifyItem{ctx: t.Context(), n: n})
		}

		time.Sleep(10 * time.Second)
		require.NoError(t, h.notifiers["fake"].Stop(t.Context()))
		require.NoError(t, h.proc.Stop(t.Context()))

		assert.Equal(t, notifyPassThrough, f.single)
		assert.Equal(t, []int{3}, f.batches)
	})
}
