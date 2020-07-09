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

type DNSRecord struct {
	ID        int64          `db:"id"         json:"-"`
	PayloadID int64          `db:"payload_id" json:"-"`
	Name      string         `db:"name"       json:"name"`
	Type      string         `db:"type"       json:"type"`
	TTL       int            `db:"ttl"        json:"ttl"`
	Values    pq.StringArray `db:"values"     json:"values"`
	CreatedAt time.Time      `db:"created_at" json:"createdAt"`
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
		name := fmt.Sprintf("%s.%s.", r.Name, origin)

		switch r.Type {

		case DNSTypeA:
			rrs = append(rrs, &dns.A{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    uint32(r.TTL),
				},
				A: net.ParseIP(v),
			})

		case DNSTypeAAAA:
			rrs = append(rrs, &dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    uint32(r.TTL),
				},
				AAAA: net.ParseIP(v),
			})

		case DNSTypeMX:
			var (
				pref uint16
				mx   string
			)

			fmt.Scanf("%d %s", &pref, &mx)

			rrs = append(rrs, &dns.MX{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    uint32(r.TTL),
				},
				Mx:         mx,
				Preference: pref,
			})

		case DNSTypeTXT:
			rrs = append(rrs, &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    uint32(r.TTL),
				},
				Txt: strings.Split(v, ","),
			})

		case DNSTypeCNAME:
			rrs = append(rrs, &dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    uint32(r.TTL),
				},
				Target: v,
			})
		}
	}

	return rrs
}
