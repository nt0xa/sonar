package server

import (
	"context"
	"log/slog"
	"net"
	"net/netip"
	"regexp"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/modules"
	"github.com/nt0xa/sonar/pkg/batcher"
	"github.com/nt0xa/sonar/pkg/geoipx"
	"github.com/nt0xa/sonar/pkg/telemetry"
	"github.com/nt0xa/sonar/pkg/workerpool"
)

type NotifyFunc func(net.Addr, []byte, map[string]any)

var (
	subdomainRegexp = regexp.MustCompile("[a-fA-F0-9]{8}")
)

const (
	// notifyWorkers bounds concurrent sends per notifier.
	notifyWorkers = 10

	// notifyPassThrough events per payload are sent in full (covers a DNS + HTTP interaction), the rest are summarized.
	notifyPassThrough = 5

	// notifyMaxBatch events per payload are held in memory per window, the rest are dropped.
	notifyMaxBatch = 1000
)

type EventsHandler struct {
	db        *database.DB
	gdb       *geoipx.DB
	log       *slog.Logger
	tel       telemetry.Telemetry
	notifiers map[string]*batcher.Batcher[notifyItem]
	proc      *workerpool.Pool[Event]
}

type Event struct {
	Event *database.Event
	Match []byte
}

// notifyItem is a notification with the ctx of the event that triggered it.
type notifyItem struct {
	ctx context.Context
	n   *modules.Notification
}

func NewEventsHandler(
	db *database.DB,
	gdb *geoipx.DB,
	log *slog.Logger,
	tel telemetry.Telemetry,
	workers int,
	capacity int,
) *EventsHandler {
	h := &EventsHandler{
		db:        db,
		gdb:       gdb,
		log:       log,
		tel:       tel,
		notifiers: make(map[string]*batcher.Batcher[notifyItem]),
	}

	h.proc = workerpool.New(workers, capacity, h.handleEvent)

	return h
}

func (h *EventsHandler) AddNotifier(name string, notifier modules.Notifier) {
	limit, burst := notifier.RateLimit()

	h.notifiers[name] = batcher.New(
		func(it notifyItem) string { return strconv.FormatInt(it.n.Payload.ID, 10) },
		func(batch []notifyItem) {
			// A window with a single event sends it in full rather than as a summary.
			if len(batch) == 1 {
				h.notify(context.Background(), batch[0].ctx, batch[0].n, notifier)
			} else {
				h.notifyBatch(batch, notifier)
			}
		},
		batcher.PassThrough(notifyPassThrough),
		batcher.MaxBatch(notifyMaxBatch),
		batcher.Workers(notifyWorkers),
		batcher.RateLimit(limit, burst),
	)
}

func (h *EventsHandler) handleEvent(ctx context.Context, e Event) {
	seen := make(map[string]struct{})

	matches := subdomainRegexp.FindAllSubmatch(e.Match, -1)
	if len(matches) == 0 {
		return
	}

	for _, m := range matches {
		d := strings.ToLower(string(m[0]))

		if _, ok := seen[d]; !ok {
			seen[d] = struct{}{}
		} else {
			continue
		}

		p, err := h.db.PayloadsGetBySubdomain(ctx, d)
		if err != nil {
			continue
		}

		e.Event.PayloadID = p.ID

		h.addGeoIPMetadata(e.Event)

		// TODO: refactor this.
		if id := getEventID(ctx); id != nil {
			e.Event.UUID = *id
		} else {
			e.Event.UUID = uuid.New()
		}

		// Store event in database
		if p.StoreEvents {
			if _, err := h.db.EventsCreate(ctx, database.EventsCreateParams{
				UUID:       e.Event.UUID,
				PayloadID:  e.Event.PayloadID,
				Protocol:   e.Event.Protocol,
				Data:       e.Event.Data,
				Meta:       e.Event.Meta,
				RemoteAddr: e.Event.RemoteAddr,
				ReceivedAt: e.Event.ReceivedAt,
			}); err != nil {
				h.log.Error("Failed to save event",
					"err", err,
					"event", e,
				)
			}
		}

		// Skip if current event protocol is muted for payload.
		if !database.ProtoCategoryContains(p.NotifyProtocols, database.ProtoToCategory(e.Event.Protocol)) {
			continue
		}

		u, err := h.db.UsersGetByID(ctx, p.UserID)
		if err != nil {
			continue
		}

		for _, c := range h.notifiers {
			// TODO: add deadline to context
			c.Push(notifyItem{
				ctx: ctx,
				n: &modules.Notification{
					User:    u,
					Payload: p,
					Event:   e.Event,
				},
			})
		}
	}
}

func (h *EventsHandler) addGeoIPMetadata(e *database.Event) {
	if h.gdb != nil {
		host, _, err := net.SplitHostPort(e.RemoteAddr)
		if err != nil {
			h.log.Error("Failed to split remote address",
				"err", err,
				"remote_addr", e.RemoteAddr,
			)
			return
		}

		ip, err := netip.ParseAddr(host)
		if err != nil {
			h.log.Error("Failed to parse remote IP",
				"err", err,
				"ip", host,
			)
			return
		}

		info, err := h.gdb.Lookup(ip)
		if err != nil {
			h.log.Error("Failed to lookup IP in GeoIP database",
				"err", err,
				"ip", ip.String(),
			)
			return
		}

		e.Meta.GeoIP = info
	}
}

func (h *EventsHandler) notify(
	ctx context.Context,
	parentCtx context.Context,
	notification *modules.Notification,
	notifier modules.Notifier,
) {
	_, span := h.tel.TraceStart(ctx, "notify",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("event.id", notification.Event.UUID.String()),
			attribute.String("notifier.name", notifier.Name()),
		),
		trace.WithLinks(trace.LinkFromContext(parentCtx)),
	)
	defer span.End()

	if err := notifier.Notify(parentCtx, notification); err != nil {
		h.log.Error("Notifier failed",
			"error", err,
			"notifier", notifier.Name(),
			"event_uuid", notification.Event.UUID.String(),
		)
	}
}

func (h *EventsHandler) notifyBatch(
	batch []notifyItem,
	notifier modules.Notifier,
) {
	ns := make([]*modules.Notification, len(batch))
	links := make([]trace.Link, len(batch))
	for i, it := range batch {
		ns[i] = it.n
		links[i] = trace.LinkFromContext(it.ctx)
	}

	ctx, span := h.tel.TraceStart(context.Background(), "notify.batch",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.Int("events.count", len(ns)),
			attribute.String("notifier.name", notifier.Name()),
		),
		trace.WithLinks(links...),
	)
	defer span.End()

	if err := notifier.NotifyBatch(ctx, ns); err != nil {
		h.log.Error("Notifier failed",
			"error", err,
			"notifier", notifier.Name(),
			"events_count", len(ns),
		)
	}
}

func (h *EventsHandler) Emit(ctx context.Context, e *database.Event, match []byte) {
	h.proc.Process(ctx, Event{Event: e, Match: match})
}

type eventIDKey struct{}

func withEventID(ctx context.Context) (context.Context, uuid.UUID) {
	id := uuid.New()
	return context.WithValue(ctx, eventIDKey{}, id), id
}

func getEventID(ctx context.Context) *uuid.UUID {
	id, ok := ctx.Value(eventIDKey{}).(uuid.UUID)
	if !ok {
		return nil
	}
	return &id
}
