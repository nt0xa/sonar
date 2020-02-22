package notifier

import (
	"net"
	"time"

	"github.com/bi-zone/sonar/internal/database"
)

type Event struct {
	Protocol   string
	Data       string
	RawData    []byte
	RemoteAddr net.Addr
	ReceivedAt time.Time
}

type Notifier interface {
	Notify(*Event, *database.Payload) error
}
