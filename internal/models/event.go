package models

import (
	"net"
	"strings"
	"time"
)

type Event struct {
	Protocol   string
	RawData    []byte
	Meta       map[string]interface{}
	RemoteAddr net.Addr
	ReceivedAt time.Time
}

func (e *Event) Proto() string {
	proto := strings.ToLower(e.Protocol)

	switch proto {

	// Change "https" to "http" because there is only
	// one category for both.
	case "https":
		return "http"

	default:
		return strings.ToLower(e.Protocol)
	}
}
