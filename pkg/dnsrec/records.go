package dnsrec

import (
	"fmt"
	"strings"
	"sync"

	"github.com/miekg/dns"

	"github.com/bi-zone/sonar/pkg/dnsutils"
)

// Records represents in memory stored DNS records.
type Records struct {
	records map[string][]dns.RR
	mu      sync.RWMutex
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
func (r *Records) Add(rr dns.RR) error {
	key := makeKey(rr.Header().Name, rr.Header().Rrtype)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.records[key]; !ok {
		r.records[key] = make([]dns.RR, 0)
	}

	r.records[key] = append(r.records[key], rr)

	return nil
}

// Del removes DNS record from Records.
func (r *Records) Del(name string, qtype uint16) error {
	key := makeKey(name, qtype)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.records[key]; !ok {
		return nil
	}

	delete(r.records, key)

	return nil
}

// Get returns DNS record from Records.
// Returns nil if no records found.
// Handles wildcards.
func (r *Records) Get(name string, qtype uint16) ([]dns.RR, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, n := range dnsutils.MakeWildcards(name) {
		key := makeKey(n, qtype)

		if res, ok := r.records[key]; ok {
			return dnsutils.ReplaceWildcards(res, name), nil
		}
	}

	return nil, nil
}

// makeKey creates string key for DNS record.
func makeKey(name string, qtype uint16) string {
	return fmt.Sprintf("%s/%s", strings.ToLower(name), dnsutils.QtypeString(qtype))
}
