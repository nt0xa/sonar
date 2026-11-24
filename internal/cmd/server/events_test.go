package server

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/google/uuid"
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
		h := NewEventsHandler(nil, nil, log, false, 1, 1)

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

func Test_EmitWithCancelledContext(t *testing.T) {
	const n = 100

	var buf bytes.Buffer
	h := NewEventsHandler(nil, nil, slog.New(slog.NewTextHandler(&buf, nil)), false, 1, n)

	// httpx emits from conn.OnClose, when req.Context() is already cancelled.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	for range n {
		h.Emit(ctx, &database.Event{Protocol: "http"}, nil)
	}

	require.NoError(t, h.proc.Stop(t.Context()))
	assert.Empty(t, buf.String())
}

func Test_EventAudit(t *testing.T) {
	dsn := os.Getenv("SONAR_DB_DSN")
	if dsn == "" {
		t.Skip("empty SONAR_DB_DSN")
	}

	_, err := database.Migrate(dsn)
	require.NoError(t, err)

	db, err := database.NewWithDSN(dsn)
	require.NoError(t, err)

	ctx := t.Context()
	subdomain := hex.EncodeToString(uuid.New().NodeID()[:4])

	u, err := db.UsersCreate(ctx, database.UsersCreateParams{Name: "audit-event-" + subdomain})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.UsersDelete(context.Background(), u.ID) })

	_, err = db.PayloadsCreate(ctx, database.PayloadsCreateParams{
		UserID:          u.ID,
		Subdomain:       subdomain,
		Name:            "audit-event",
		NotifyProtocols: []string{},
	})
	require.NoError(t, err)

	h := NewEventsHandler(db, nil, slog.New(slog.DiscardHandler), true, 1, 1)
	h.handleEvent(Event{
		ctx:   ctx,
		Event: &database.Event{Protocol: "dns", Data: [][]byte{}, RemoteAddr: "127.0.0.1:53", ReceivedAt: time.Now()},
		Match: []byte(subdomain),
	})

	recs, err := db.AuditRecordsList(ctx, database.AuditRecordsListParams{
		ResourceType: "event",
		Action:       "create",
		PageLimit:    1,
	})
	require.NoError(t, err)
	require.Len(t, recs, 1)

	rec := recs[0]
	assert.Equal(t, database.AuditRecordSourceTypeSystem, rec.Source)
	assert.Nil(t, rec.ActorID)

	var resource map[string]any
	require.NoError(t, json.Unmarshal(rec.Resource, &resource))
	assert.Equal(t, subdomain, resource["payloadSubdomain"])
	assert.Equal(t, "dns", resource["protocol"])
	assert.Equal(t, "127.0.0.1:53", resource["remoteAddress"])
}
