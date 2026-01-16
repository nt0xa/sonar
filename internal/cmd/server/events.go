package server

import (
	"context"
	"log/slog"
	"net"
	"net/netip"
	"regexp"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"
	"github.com/nt0xa/sonar/internal/cache"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/modules"
	"github.com/nt0xa/sonar/internal/utils"
	"github.com/nt0xa/sonar/pkg/geoipx"
	"github.com/nt0xa/sonar/pkg/telemetry"
)

type NotifyFunc func(net.Addr, []byte, map[string]any)

var (
	subdomainRegexp = regexp.MustCompile("[a-fA-F0-9]{8}")
)

type EventsHandler struct {
	db           *database.DB
	gdb          *geoipx.DB
	log          *slog.Logger
	tel          telemetry.Telemetry
	cache        cache.Cache
	workersCount int
	workersWg    sync.WaitGroup
	events       chan eventWithContext
	notifiers    map[string]modules.Notifier
}

type eventWithContext struct {
	ctx   context.Context
	event *models.Event
}

func NewEventsHandler(
	db *database.DB,
	gdb *geoipx.DB,
	log *slog.Logger,
	tel telemetry.Telemetry,
	cache cache.Cache,
	workers int,
	capacity int,
) *EventsHandler {
	return &EventsHandler{
		db:           db,
		gdb:          gdb,
		log:          log,
		tel:          tel,
		cache:        cache,
		workersCount: workers,
		events:       make(chan eventWithContext, capacity),
		notifiers:    make(map[string]modules.Notifier),
	}
}

func (h *EventsHandler) AddNotifier(name string, notifier modules.Notifier) {
	h.notifiers[name] = notifier
}

func (h *EventsHandler) Start() error {
	for i := 0; i < h.workersCount; i++ {
		h.workersWg.Add(1)
		go h.worker(i)
	}

	return nil
}

func (h *EventsHandler) worker(id int) {
	defer h.workersWg.Done()

	for e := range h.events {
		ctx := context.Background()

		if id := getEventID(e.ctx); id != nil {
			e.event.UUID = *id
		} else {
			e.event.UUID = uuid.New()
			h.log.Warn("Event ID not found in context, generating new one")
		}

		ctx, span := h.tel.TraceStart(ctx, "event",
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(
				attribute.String("event.id", e.event.UUID.String()),
				attribute.Int("event.worker.id", id),
				attribute.String("event.protocol", e.event.Protocol.Name),
			),
			trace.WithLinks(trace.LinkFromContext(e.ctx)),
		)
		h.handleEvent(ctx, e.event)
		span.End()
	}
}

func (h *EventsHandler) handleEvent(ctx context.Context, e *models.Event) {
	seen := make(map[string]struct{})

	matches := subdomainRegexp.FindAllSubmatch(e.R, -1)
	if len(matches) == 0 {
		return
	}

	for _, m := range matches {
		d := strings.ToLower(string(m[0]))

		if !h.cache.SubdomainExists(d) {
			continue
		}

		if _, ok := seen[d]; !ok {
			seen[d] = struct{}{}
		} else {
			continue
		}

		p, err := h.db.PayloadsGetBySubdomain(ctx, d)
		if err != nil {
			continue
		}

		e.PayloadID = p.ID

		h.addGeoIPMetadata(e)

		// Store event in database
		if p.StoreEvents {
			if err := h.db.EventsCreate(ctx, e); err != nil {
				h.log.Error("Failed to save event",
					"err", err,
					"event", e,
				)
			}
		}

		// Skip if current event protocol is muted for payload.
		if !p.NotifyProtocols.Contains(e.Protocol.Category()) {
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
				Event:   e,
			}, n)
		}
	}
}

func (h *EventsHandler) addGeoIPMetadata(e *models.Event) {
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

		e.Meta.GeoIP = &models.GeoIPMeta{
			City:         info.City,
			Subdivisions: info.Subdivisions,
			Country: models.GeoIPCountry{
				Name:      info.Country.Name,
				ISOCode:   info.Country.ISOCode,
				FlagEmoji: info.Country.FlagEmoji,
			},
			ASN: models.GeoIPASN{
				Number: info.ASN.Number,
				Org:    info.ASN.Org,
			},
		}
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

func (h *EventsHandler) Emit(ctx context.Context, e *models.Event) {
	h.events <- eventWithContext{ctx: ctx, event: e}
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
