package server

import (
	"context"
	"log/slog"
	"net"
	"net/netip"
	"regexp"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"
	"github.com/nt0xa/sonar/internal/cache"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/modules"
	"github.com/nt0xa/sonar/pkg/geoipx"
	"github.com/nt0xa/sonar/pkg/telemetry"
	"github.com/nt0xa/sonar/pkg/workerpool"
)

type NotifyFunc func(net.Addr, []byte, map[string]any)

var (
	subdomainRegexp = regexp.MustCompile("[a-fA-F0-9]{8}")
)

type EventsHandler struct {
	db        *database.DB
	gdb       *geoipx.DB
	log       *slog.Logger
	tel       telemetry.Telemetry
	cache     cache.Cache
	notifiers map[string]modules.Notifier
	proc      *workerpool.Processor[Event]
}

type Event struct {
	Event *database.Event
	Match []byte
}

func NewEventsHandler(
	db *database.DB,
	gdb *geoipx.DB,
	log *slog.Logger,
	tel telemetry.Telemetry,
	cache cache.Cache,
	workers int,
	capacity int,
) (*EventsHandler, error) {
	h := &EventsHandler{
		db:        db,
		gdb:       gdb,
		log:       log,
		tel:       tel,
		cache:     cache,
		notifiers: make(map[string]modules.Notifier),
	}

	proc, err := workerpool.NewProcessor(workers, capacity, h.handleEvent)
	if err != nil {
		return nil, err
	}

	h.proc = proc

	return h, err
}

func (h *EventsHandler) AddNotifier(name string, notifier modules.Notifier) {
	h.notifiers[name] = notifier
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

		for _, n := range h.notifiers {
			// TODO: add deadline to context
			go h.notify(context.Background(), ctx, &modules.Notification{
				User:    u,
				Payload: p,
				Event:   e.Event,
			}, n)
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
