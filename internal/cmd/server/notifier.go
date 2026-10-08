package server

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/nt0xa/sonar/internal/modules"
	"github.com/nt0xa/sonar/pkg/batcher"
	"github.com/nt0xa/sonar/pkg/telemetry"
	"github.com/nt0xa/sonar/pkg/workerpool"
)

// Notifier batches notifications per payload and sends them to a notifier
// from a pool of workers within the notifier's rate limit.
type Notifier struct {
	notifier modules.Notifier
	log      *slog.Logger
	tel      telemetry.Telemetry
	batcher  *batcher.Batcher[notifyItem, int64]
	pool     *workerpool.Pool[[]notifyItem]
}

// notifyItem is a notification with the ctx of the event that triggered it.
type notifyItem struct {
	ctx context.Context
	n   *modules.Notification
}

func NewNotifier(
	notifier modules.Notifier,
	log *slog.Logger,
	tel telemetry.Telemetry,
	workers int,
	opts ...batcher.Option,
) *Notifier {
	limit, burst := notifier.RateLimit()

	n := &Notifier{
		notifier: notifier,
		log:      log,
		tel:      tel,
		batcher:  batcher.New(func(it notifyItem) int64 { return it.n.Payload.ID }, opts...),
	}

	n.pool = workerpool.New(n.send,
		workerpool.WithWorkers(workers),
		workerpool.WithRateLimit(limit, burst),
	)

	go n.forward()

	return n
}

// Add queues notification; ctx is the ctx of the event that triggered it.
func (n *Notifier) Add(ctx context.Context, notification *modules.Notification) bool {
	return n.batcher.Add(notifyItem{ctx: ctx, n: notification})
}

// Close flushes pending batches; the pool stops once they are sent.
func (n *Notifier) Close() {
	n.batcher.Close()
}

func (n *Notifier) forward() {
	for batch := range n.batcher.Batches() {
		if err := n.pool.Submit(context.Background(), batch); err != nil {
			n.log.Error("Failed to submit notifications", "err", err, "notifier", n.notifier.Name())
		}
	}
	_ = n.pool.Stop(context.Background())
}

func (n *Notifier) send(ctx context.Context, batch []notifyItem) {
	// A window with a single event sends it in full rather than as a summary.
	if len(batch) == 1 {
		n.notify(ctx, batch[0].ctx, batch[0].n)
	} else {
		n.notifyBatch(batch)
	}
}

func (n *Notifier) notify(
	ctx context.Context,
	parentCtx context.Context,
	notification *modules.Notification,
) {
	_, span := n.tel.TraceStart(ctx, "notify",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("event.id", notification.Event.UUID.String()),
			attribute.String("notifier.name", n.notifier.Name()),
		),
		trace.WithLinks(trace.LinkFromContext(parentCtx)),
	)
	defer span.End()

	if err := n.notifier.Notify(parentCtx, notification); err != nil {
		n.log.Error("Notifier failed",
			"error", err,
			"notifier", n.notifier.Name(),
			"event_uuid", notification.Event.UUID.String(),
		)
	}
}

func (n *Notifier) notifyBatch(batch []notifyItem) {
	ns := make([]*modules.Notification, len(batch))
	for i, it := range batch {
		ns[i] = it.n
	}

	if err := n.notifier.NotifyBatch(context.Background(), ns); err != nil {
		n.log.Error("Notifier failed",
			"error", err,
			"notifier", n.notifier.Name(),
			"events_count", len(ns),
		)
	}
}
