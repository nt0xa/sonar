package auditsvc

import (
	"time"
	"uuid"

	"github.com/nt0xa/sonar/internal/database"
)

// Event is the audit resource for a captured interaction event.
type Event struct {
	UUID             uuid.UUID           `json:"uuid"`
	PayloadSubdomain string              `json:"payloadSubdomain"`
	Protocol         string              `json:"protocol"`
	Data             [][]byte            `json:"data"`
	Meta             database.EventsMeta `json:"meta"`
	RemoteAddr       string              `json:"remoteAddress"`
	ReceivedAt       time.Time           `json:"receivedAt"`
}

// NewEvent returns the audit resource for event e captured for payload p.
func NewEvent(p *database.Payload, e *database.Event) Event {
	return Event{
		UUID:             e.UUID,
		PayloadSubdomain: p.Subdomain,
		Protocol:         e.Protocol,
		Data:             e.Data,
		Meta:             e.Meta,
		RemoteAddr:       e.RemoteAddr,
		ReceivedAt:       e.ReceivedAt,
	}
}
