package server

import (
	"context"
	"log/slog"
	"net"
	"regexp"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/nt0xa/sonar/internal/cache"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/modules"
	"github.com/nt0xa/sonar/pkg/telemetry"
)

type NotifyFunc func(net.Addr, []byte, map[string]interface{})

var (
	subdomainRegexp = regexp.MustCompile("[a-fA-F0-9]{8}")
)

type EventsHandler struct {
	db           *database.DB
	log          *slog.Logger
	tel          telemetry.Telemetry
	cache        cache.Cache
	workersCount int
	workersWg    sync.WaitGroup
	events       chan *models.Event
	notifiers    map[string]modules.Notifier
}

func NewEventsHandler(
	db *database.DB,
	log *slog.Logger,
	tel telemetry.Telemetry,
	cache cache.Cache,
	workers int,
	capacity int,
) *EventsHandler {
	return &EventsHandler{
		db:           db,
		log:          log,
		tel:          tel,
		cache:        cache,
		workersCount: workers,
		events:       make(chan *models.Event, capacity),
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

	for event := range h.events {
		ctx, span := h.tel.TraceStart(context.Background(), "event",
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(
				attribute.Int("event.worker.id", id),
				attribute.String("event.protocol", event.Protocol.Name),
			),
		)
		h.handleEvent(ctx, event)
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

		// Store event in database
		if p.StoreEvents {
			if err := h.db.EventsCreate(ctx, e); err != nil {
				h.log.Error("Failed to save event",
					"err", err,
					"event", e,
				)
				continue
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
			go func() {
				ctx, span := h.tel.TraceStart(ctx, "notifier.Notify",
					trace.WithSpanKind(trace.SpanKindInternal),
					trace.WithAttributes(
						attribute.String("notifier.name", n.Name()),
					),
				)
				defer span.End()

				if err := n.Notify(ctx, &modules.Notification{User: u, Payload: p, Event: e}); err != nil {
					h.log.Error("Notifier failed",
						"error", err,
						"payload_id", p.ID,
						"user_id", u.ID,
					)
				}
			}()
		}
	}
}

func (h *EventsHandler) Emit(e *models.Event) {
	h.events <- e
}
