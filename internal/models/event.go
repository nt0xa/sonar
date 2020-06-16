package models

import (
	"net"
	"time"
)

type Event struct {
	Protocol   string
	Data       string
	RawData    []byte
	RemoteAddr net.Addr
	ReceivedAt time.Time
}
