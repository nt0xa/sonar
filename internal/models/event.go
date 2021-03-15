package models

import (
	"net"
	"time"
)

type Event struct {
	Protocol   Proto
	RawData    []byte
	Meta       map[string]interface{}
	RemoteAddr net.Addr
	ReceivedAt time.Time
}

func (e *Event) ProtoCategory() ProtoCategory {
	return ProtoToCategory(e.Protocol)
}
