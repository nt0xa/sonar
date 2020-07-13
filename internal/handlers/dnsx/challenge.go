package dnsx

import (
	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/miekg/dns"
)

// Present allows DNSX to satisfy challenge.Provider interface
func (h *DNSX) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	h.addStatic(dns.RR(&dns.TXT{
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

// CleanUp allows DNSX to satisfy challenge.Provider interface
func (h *DNSX) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	h.delStatic(fqdn, dns.TypeTXT)

	return nil
}
