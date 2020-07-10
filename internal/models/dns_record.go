package models

import (
	"fmt"
	"net"
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

const (
	DNSStrategyDefault    = "default"
	DNSStrategyRoundRobin = "round-robin"
	DNSStrategyRebind     = "rebind"
)

type DNSRecord struct {
	ID             int64          `db:"id"               json:"-"`
	PayloadID      int64          `db:"payload_id"       json:"-"`
	Name           string         `db:"name"             json:"name"`
	Type           string         `db:"type"             json:"type"`
	TTL            int            `db:"ttl"              json:"ttl"`
	Values         pq.StringArray `db:"values"           json:"values"`
	Strategy       string         `db:"strategy"         json:"strategy"`
	LastAnswer     pq.StringArray `db:"last_answer"      json:"lastAnswer"`
	LastAccessedAt *time.Time     `db:"last_accessed_at" json:"lastAccessedAt"`
	CreatedAt      time.Time      `db:"created_at"       json:"createdAt"`
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
	}

	return 0
}

func (r *DNSRecord) RRs(origin string) []dns.RR {
	rrs := make([]dns.RR, 0)
	for _, v := range r.Values {
		rrs = append(rrs, DNSStringToRR(v, r.Type, r.Name, origin, r.TTL))
	}
	return rrs
}

func (r *DNSRecord) LastAnswerRRs(origin string) []dns.RR {
	rrs := make([]dns.RR, 0)
	for _, v := range r.LastAnswer {
		rrs = append(rrs, DNSStringToRR(v, r.Type, r.Name, origin, r.TTL))
	}
	return rrs
}

func DNSStringToRR(value, typ, name, origin string, ttl int) dns.RR {
	fqdn := fmt.Sprintf("%s.%s.", name, origin)

	switch typ {

	case DNSTypeA:
		return &dns.A{
			Hdr: dns.RR_Header{
				Name:   fqdn,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    uint32(ttl),
			},
			A: net.ParseIP(value),
		}

	case DNSTypeAAAA:
		return &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   fqdn,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    uint32(ttl),
			},
			AAAA: net.ParseIP(value),
		}

	case DNSTypeMX:
		var (
			pref uint16
			mx   string
		)

		fmt.Sscanf(value, "%d %s", &pref, &mx)

		return &dns.MX{
			Hdr: dns.RR_Header{
				Name:   fqdn,
				Rrtype: dns.TypeMX,
				Class:  dns.ClassINET,
				Ttl:    uint32(ttl),
			},
			Mx:         mx,
			Preference: pref,
		}

	case DNSTypeTXT:
		return &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   fqdn,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    uint32(ttl),
			},
			Txt: strings.Split(value, ","),
		}

	case DNSTypeCNAME:
		return &dns.CNAME{
			Hdr: dns.RR_Header{
				Name:   fqdn,
				Rrtype: dns.TypeCNAME,
				Class:  dns.ClassINET,
				Ttl:    uint32(ttl),
			},
			Target: value,
		}
	}

	return nil
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

	return ""
}
