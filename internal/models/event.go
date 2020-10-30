package models

import (
	"net"
	"time"
)

type Event struct {
	Protocol   string
	Data       string
	RawData    []byte
	Meta       map[string]interface{}
	RemoteAddr net.Addr
	ReceivedAt time.Time
}
