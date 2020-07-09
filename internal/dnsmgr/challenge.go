package dnsmgr

import (
	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/miekg/dns"
)

// Present allows DNSMgr to satisfy challenge.Provider interface
func (mgr *DNSMgr) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	mgr.addStatic(dns.RR(&dns.TXT{
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

// CleanUp allows DNSMgr to satisfy challenge.Provider interface
func (mgr *DNSMgr) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	mgr.delStatic(fqdn, dns.TypeTXT)

	return nil
}
