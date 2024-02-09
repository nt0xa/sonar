package server

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"

	"github.com/russtone/sonar/internal/cache"
	"github.com/russtone/sonar/internal/database"
	"github.com/russtone/sonar/internal/database/models"
	"github.com/russtone/sonar/internal/modules"
)

type NotifyFunc func(net.Addr, []byte, map[string]interface{})

var (
	subdomainRegexp = regexp.MustCompile("[a-fA-F0-9]{8}")
)

type EventsHandler struct {
	db           *database.DB
	cache        cache.Cache
	workersCount int
	workersWg    sync.WaitGroup
	events       chan *models.Event
	notifiers    map[string]modules.Notifier
}

func NewEventsHandler(db *database.DB, cache cache.Cache, workers int, capacity int) *EventsHandler {
	return &EventsHandler{
		db:           db,
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

	for e := range h.events {

		seen := make(map[string]struct{})

		matches := subdomainRegexp.FindAllSubmatch(e.R, -1)
		if len(matches) == 0 {
			continue
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

			p, err := h.db.PayloadsGetBySubdomain(d)
			if err != nil {
				continue
			}

			e.PayloadID = p.ID

			// Store event in database
			if p.StoreEvents {
				if err := h.db.EventsCreate(e); err != nil {
					fmt.Println(err)
					continue
				}
			}

			// Skip if current event protocol is muted for payload.
			if !p.NotifyProtocols.Contains(e.Protocol.Category()) {
				continue
			}

			u, err := h.db.UsersGetByID(p.UserID)
			if err != nil {
				continue
			}

			for _, n := range h.notifiers {

				if err := n.Notify(&modules.Notification{User: u, Payload: p, Event: e}); err != nil {
					continue
				}
			}

		}
	}
}

func (h *EventsHandler) Emit(e *models.Event) {
	h.events <- e
}
