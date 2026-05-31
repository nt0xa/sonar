package types

import (
	"time"
)

type Payload struct {
	Name            string
	Subdomain       string
	NotifyProtocols []string
	StoreEvents     bool
	CreatedAt       time.Time
}
