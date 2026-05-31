package types

import (
	"time"
)

type DNSRecord struct {
	Index            int64
	PayloadSubdomain string
	Name             string
	Type             string
	TTL              int
	Values           []string
	Strategy         string
	CreatedAt        time.Time
}
