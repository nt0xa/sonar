package dnsrec

import (
	"fmt"
	"strings"
	"sync"

	"github.com/miekg/dns"

	"github.com/bi-zone/sonar/internal/dnsx/dnsutils"
)

// Records represents in memory stored DNS records.
type Records struct {
	records map[string][]dns.RR
	mu      sync.Mutex
}

// New returns new initialized Records instance.
func New(rrs []dns.RR) *Records {
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

	if res, ok := r.records[key]; !ok {
		return nil
	} else {
		return res
	}
}

// makeKey creates string key for DNS record.
func makeKey(name string, qtype uint16) string {
	return fmt.Sprintf("%s/%s", strings.ToLower(name), dnsutils.QtypeString(qtype))
}
