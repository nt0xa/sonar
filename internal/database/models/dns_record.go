package models

import (
	"time"

	"github.com/lib/pq"
	"github.com/miekg/dns"
)

const (
	DNSTypeA     = "A"
	DNSTypeAAAA  = "AAAA"
	DNSTypeMX    = "MX"
	DNSTypeTXT   = "TXT"
	DNSTypeCNAME = "CNAME"
	DNSTypeNS    = "NS"
)

var DNSTypesAll = []string{
	DNSTypeA,
	DNSTypeAAAA,
	DNSTypeMX,
	DNSTypeTXT,
	DNSTypeCNAME,
	DNSTypeNS,
}

const (
	DNSStrategyAll        = "all"
	DNSStrategyRoundRobin = "round-robin"
	DNSStrategyRebind     = "rebind"
)

var DNSStrategiesAll = []string{
	DNSStrategyAll,
	DNSStrategyRoundRobin,
	DNSStrategyRebind,
}

type DNSRecord struct {
	Index          int64          `db:"index"`
	ID             int64          `db:"id"`
	PayloadID      int64          `db:"payload_id"`
	Name           string         `db:"name"`
	Type           string         `db:"type"`
	TTL            int            `db:"ttl"`
	Values         pq.StringArray `db:"values"`
	Strategy       string         `db:"strategy"`
	LastAnswer     pq.StringArray `db:"last_answer"`
	LastAccessedAt *time.Time     `db:"last_accessed_at"`
	CreatedAt      time.Time      `db:"created_at"`
}

func (r *DNSRecord) Qtype() uint16 {
	switch r.Type {
	case DNSTypeA:
		return dns.TypeA
	case DNSTypeAAAA:
		return dns.TypeAAAA
	case DNSTypeMX:
		return dns.TypeMX
	case DNSTypeTXT:
		return dns.TypeTXT
	case DNSTypeCNAME:
		return dns.TypeCNAME
	case DNSTypeNS:
		return dns.TypeNS
	}
	panic("unsupported dns query type")
}
