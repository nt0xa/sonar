package dnsutils

import (
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/miekg/dns"
)

// ParseRecords parses DNS records from string.
func ParseRecords(s string, origin string) ([]dns.RR, error) {
	ss := s

	// We need a closing newline
	if len(s) > 0 && s[len(s)-1] != '\n' {
		ss += "\n"
	}

	return ParseRecordsFile(strings.NewReader(s+"\n"), origin)
}

// ParseRecordsFile parses DNS records from file.
func ParseRecordsFile(rdr io.Reader, origin string) ([]dns.RR, error) {
	rrs := make([]dns.RR, 0)
	zp := dns.NewZoneParser(rdr, origin, "")

	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		rrs = append(rrs, rr)
	}

	if err := zp.Err(); err != nil {
		return nil, err
	}

	return rrs, nil
}

// Must is a helper that wraps a call to a function returning ([]dns.RR, error)
// and panics if the error is non-nil.
func Must(rrs []dns.RR, err error) []dns.RR {
	if err != nil {
		panic(err)
	}
	return rrs
}

// NewRR creates dns.RR using provided parameters.
func NewRR(name string, qtype uint16, ttl int, value string) dns.RR {
	switch qtype {

	case dns.TypeA:
		return &dns.A{
			Hdr: dns.RR_Header{
				Name:   name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    uint32(ttl),
			},
			A: net.ParseIP(value),
		}

	case dns.TypeAAAA:
		return &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   name,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    uint32(ttl),
			},
			AAAA: net.ParseIP(value),
		}

	case dns.TypeMX:
		var (
			pref uint16
			mx   string
		)

		_, _ = fmt.Sscanf(value, "%d %s", &pref, &mx)

		return &dns.MX{
			Hdr: dns.RR_Header{
				Name:   name,
				Rrtype: dns.TypeMX,
				Class:  dns.ClassINET,
				Ttl:    uint32(ttl),
			},
			Mx:         mx,
			Preference: pref,
		}

	case dns.TypeTXT:
		return &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   name,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    uint32(ttl),
			},
			Txt: strings.Split(value, ","),
		}

	case dns.TypeCNAME:
		return &dns.CNAME{
			Hdr: dns.RR_Header{
				Name:   name,
				Rrtype: dns.TypeCNAME,
				Class:  dns.ClassINET,
				Ttl:    uint32(ttl),
			},
			Target: value,
		}
	}

	return nil
}

// NewRRs creates array of dns.RR using provided parameters.
func NewRRs(name string, qtype uint16, ttl int, values []string) []dns.RR {
	rrs := make([]dns.RR, 0)

	for _, value := range values {
		rrs = append(rrs, NewRR(name, qtype, ttl, value))
	}

	return rrs
}

// QtypeString return string representation of uint16 DNS query type.
func QtypeString(qtype uint16) string {
	return dns.Type(qtype).String()
}

// RRToString returns string representation of dns.RR value.
func RRToString(rr dns.RR) string {

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

// RRsToStings returns representaion of []dns.RR as []string.
func RRsToStrings(rrs []dns.RR) []string {
	res := make([]string, 0)
	for _, rr := range rrs {
		res = append(res, RRToString(rr))
	}
	return res
}

func MakeWildcard(fqdn string) string {
	split := strings.SplitAfterN(fqdn, ".", 2)
	split[0] = "*"
	return strings.Join(split, ".")
}

func MakeWildcards(fqdn string) []string {
	res := make([]string, 0)
	for off, end := 0, false; !end; off, end = dns.NextLabel(fqdn, off) {
		res = append(res, MakeWildcard(fqdn[off:]))
	}
	return res
}
