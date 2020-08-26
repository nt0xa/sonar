package models

import (
	"time"

	"github.com/lib/pq"
)

const (
	PayloadProtocolDNS  = "dns"
	PayloadProtocolHTTP = "http"
	PayloadProtocolSMTP = "smtp"
)

var PayloadProtocolsAll = []string{
	PayloadProtocolDNS,
	PayloadProtocolHTTP,
	PayloadProtocolSMTP,
}

type Payload struct {
	ID              int64          `db:"id"`
	UserID          int64          `db:"user_id"`
	Subdomain       string         `db:"subdomain"`
	Name            string         `db:"name"`
	NotifyProtocols pq.StringArray `db:"notify_protocols"`
	CreatedAt       time.Time      `db:"created_at"`
}
