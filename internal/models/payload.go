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
	ID              int64          `db:"id"                json:"-"`
	UserID          int64          `db:"user_id"           json:"-"`
	Subdomain       string         `db:"subdomain"         json:"subdomain"`
	Name            string         `db:"name"              json:"name"`
	NotifyProtocols pq.StringArray `db:"notify_protocols"  json:"notifyProtocols,omitempty"`
	CreatedAt       time.Time      `db:"created_at"        json:"createdAt"`
}
