package dnsmgr

import (
	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/miekg/dns"
)

// Present allows DNSMgr to satisfy challenge.Provider interface
func (mgr *DNSMgr) Present(domain, token, keyAuth string) error {
	mgr.Lock()
	defer mgr.Unlock()

	fqdn, value := dns01.GetRecord(domain, keyAuth)

	mgr.staticRecords.add(dns.RR(&dns.TXT{
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
	mgr.Lock()
	defer mgr.Unlock()

	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	mgr.staticRecords.delByName(fqdn)

	return nil
}
