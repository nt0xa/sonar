package dnsrec

import (
	"github.com/miekg/dns"

	"github.com/bi-zone/sonar/internal/protocols/dnsx"
)

// Ensure dnsx.Handler interface is implemented.
var _ dnsx.Handler = &Records{}

// ServeDNS allows Records to implement dnsx.Handler interface.
func (r *Records) ServeDNS(name string, qtype uint16) ([]dns.RR, error) {
	return r.Get(name, qtype), nil
}
