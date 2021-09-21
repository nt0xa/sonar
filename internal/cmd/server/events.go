package server

import (
	"database/sql"
	"net"
	"regexp"
	"strings"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/database/models"
)

type NotifyFunc func(net.Addr, []byte, map[string]interface{})

var (
	subdomainRegexp = regexp.MustCompile("[a-fA-F0-9]{8}")
)

type EventsHandler struct {
	db        *database.DB
	events    chan *models.Event
	notifiers map[string]Notifier
}

func NewEventsHandler(db *database.DB, capacity int) *EventsHandler {
	return &EventsHandler{
		db:        db,
		events:    make(chan *models.Event, capacity),
		notifiers: make(map[string]Notifier),
	}
}

func (h *EventsHandler) AddNotifier(name string, notifier Notifier) {
	h.notifiers[name] = notifier
}

func (h *EventsHandler) Start() error {
	for e := range h.events {

		seen := make(map[string]struct{})

		matches := subdomainRegexp.FindAllSubmatch(e.RW, -1)
		if len(matches) == 0 {
			continue
		}

		for _, m := range matches {
			d := strings.ToLower(string(m[0]))

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
			if p.StoreEvents > 0 {
				if err := h.db.EventsCreate(e); err != nil {
					continue
				}

				// Delete out of limit events
				if err := h.db.EventsDeleteOutOfLimit(p.ID, p.StoreEvents); err != nil && err != sql.ErrNoRows {
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
				if err := n.Notify(u, p, e); err != nil {
					continue
				}
			}

		}
	}

	return nil
}

func (h *EventsHandler) Emit(e *models.Event) {
	h.events <- e
}
