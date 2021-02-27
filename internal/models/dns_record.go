package models

import (
	"fmt"
	"strings"
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
)

var DNSTypesAll = []string{
	DNSTypeA,
	DNSTypeAAAA,
	DNSTypeMX,
	DNSTypeTXT,
	DNSTypeCNAME,
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
	return DNSQtype(r.Type)
}

func DNSQtype(typ string) uint16 {
	switch typ {
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
	}
	panic("unsupported dns query type")
}

func DNSType(qtype uint16) string {
	switch qtype {
	case dns.TypeA:
		return DNSTypeA
	case dns.TypeAAAA:
		return DNSTypeAAAA
	case dns.TypeMX:
		return DNSTypeMX
	case dns.TypeTXT:
		return DNSTypeTXT
	case dns.TypeCNAME:
		return DNSTypeCNAME
	}
	panic("unsupported dns query type")
}

func DNSRRToString(rr dns.RR) string {

	switch r := rr.(type) {
	case *dns.A:
		return r.A.String()
	case *dns.AAAA:
		return r.AAAA.String()
	case *dns.MX:
		return fmt.Sprintf("%d %s", r.Preference, r.Mx)
	case *dns.TXT:
		return strings.Join(r.Txt, ",")
	case *dns.CNAME:
		return r.Target
	}

	panic("unsupported dns record type")
}
