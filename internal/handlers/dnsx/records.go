package dnsx

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/miekg/dns"

	"github.com/bi-zone/sonar/internal/utils/tpl"
)

var recordsTpl = tpl.MustParse(`
@ IN 60 NS ns1
* IN 60 NS ns1
@ IN 60 NS ns2
* IN 60 NS ns2

{{ if .To4 -}}
@ IN 60 A {{ . }}
* IN 60 A {{ . }}
@ IN 60 AAAA ::ffff:{{ . }}
* IN 60 AAAA ::ffff:{{ . }}
{{- else -}}
@ IN 60 AAAA {{ . }}
* IN 60 AAAA {{ . }}
{{- end }}

@ 60 IN MX 10 mx
* 60 IN MX 10 mx

@ 60 IN CAA 60 issue "letsencrypt.org"
`)

// Records represents in memory stored DNS records.
type Records struct {
	records map[string][]dns.RR
	mu      sync.Mutex
}

// Records must implement Finder interface.
var _ Finder = &Records{}

// NewRecords returns new initialized Records instance.
func NewRecords(rrs []dns.RR) *Records {
	rec := &Records{
		records: make(map[string][]dns.RR),
	}

	for _, rr := range rrs {
		rec.Add(rr)
	}

	return rec
}

// Add adds new DNS record to Records.
func (r *Records) Add(rr dns.RR) {
	key := makeKey(rr.Header().Name, rr.Header().Rrtype)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.records[key]; !ok {
		r.records[key] = make([]dns.RR, 0)
	}

	r.records[key] = append(r.records[key], rr)
}

// Del removes DNS record from Records.
func (r *Records) Del(name string, qtype uint16) {
	key := makeKey(name, qtype)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.records[key]; !ok {
		return
	}

	delete(r.records, key)
}

// Get returns DNS record from Records.
// Returns nil if no records found.
func (r *Records) Get(name string, qtype uint16) []dns.RR {
	key := makeKey(name, qtype)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.records[key]; !ok {
		return nil
	}

	return r.records[key]
}

// Find allows Records to implement Finder interface.
func (r *Records) Find(name string, qtype uint16) ([]dns.RR, error) {
	return r.Get(name, qtype), nil
}

// Present allows Records to satisfy challenge.Provider interface
func (r *Records) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	r.Add(dns.RR(&dns.TXT{
		Hdr: dns.RR_Header{
			Name:   fqdn,
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    60,
		},
		Txt: []string{value},
	}))

	return nil
}

// CleanUp allows Records to satisfy challenge.Provider interface
func (r *Records) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	r.Del(fqdn, dns.TypeTXT)

	return nil
}

// ParseRecords parses DNS records from string.
func ParseRecords(s string, origin string) ([]dns.RR, error) {
	ss := s

	// We need a closing newline
	if len(s) > 0 && s[len(s)-1] != '\n' {
		ss += "\n"
	}

	return ParseRecordsFile(strings.NewReader(s + "\n"), origin)
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

// DefaultRecords returns default DNS records.
func DefaultRecords(origin string, ip net.IP) (*Records, error) {
	s, err := tpl.RenderToString(recordsTpl, ip)
	if err != nil {
		return nil, err
	}

	rrs, err := ParseRecords(s, origin)
	if err != nil {
		return nil, err
	}

	return NewRecords(rrs), nil
}
