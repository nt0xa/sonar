// Package dnschal provides wrapper for dnsutils.Records which
// implements lego challenge.Provider interface.
package dnschal

import (
	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/miekg/dns"

	"github.com/bi-zone/sonar/internal/dnsx/dnsrec"
)

type Provider struct {
	*dnsrec.Records
}

// Present allows Records to satisfy challenge.Provider interface
func (p *Provider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	p.Add(dns.RR(&dns.TXT{
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
func (p *Provider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	p.Del(fqdn, dns.TypeTXT)

	return nil
}
