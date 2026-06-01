package types

import (
	"time"
)

//go:generate go-enum --ptr --names --values

// ENUM(A, AAAA, MX, TXT, CNAME, NS, CAA)
type DNSRecordType string

// ENUM(all, round-robin, rebind)
type DNSRecordStrategy string

type DNSRecord struct {
	Index            int64
	PayloadSubdomain string
	Name             string
	Type             DNSRecordType
	TTL              int
	Values           []string
	Strategy         DNSRecordStrategy
	CreatedAt        time.Time
}
