package models

import (
	"net"
	"time"
)

type Event struct {
	Protocol   Proto `db:"protocol"`
	Log        []byte
	To         []byte
	From       []byte
	Meta       map[string]interface{}
	RemoteAddr net.Addr
	ReceivedAt time.Time
}
