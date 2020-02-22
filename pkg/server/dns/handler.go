package dns

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/miekg/dns"

	"github.com/bi-zone/sonar/pkg/server"
)

type handler struct {
	ip          net.IP
	domain      string
	ttl         uint32
	soa         uint32
	handlerFunc server.HandlerFunc

	// For Let's Encrypt challenges
	txtRecords sync.Map
}

// ServeDNS allows handler to satisfy dns.Handler interface
func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}

	msg.SetReply(r)

	switch r.Question[0].Qtype {

	case dns.TypeAAAA:
		msg.Answer = append(msg.Answer, &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   msg.Question[0].Name,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    h.ttl,
			},
			AAAA: h.ip,
		})

	case dns.TypeA:
		msg.Answer = append(msg.Answer, &dns.A{
			Hdr: dns.RR_Header{
				Name:   msg.Question[0].Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    h.ttl,
			},
			A: h.ip,
		})

	case dns.TypeMX:
		msg.Answer = append(msg.Answer, &dns.MX{
			Hdr: dns.RR_Header{
				Name:   msg.Question[0].Name,
				Rrtype: dns.TypeMX,
				Class:  dns.ClassINET,
				Ttl:    h.ttl,
			},
			Preference: 10,
			Mx:         fmt.Sprintf("mx.%s.", h.domain),
		})

	case dns.TypeTXT:

		if values, ok := h.txtRecords.Load(strings.ToLower(msg.Question[0].Name)); ok {
			records := values.([]string)
			for _, value := range records {
				msg.Answer = append(msg.Answer, &dns.TXT{
					Hdr: dns.RR_Header{
						Name:   msg.Question[0].Name,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    h.ttl,
					},
					Txt: []string{value},
				})
			}
		}

	case dns.TypeSOA:
		msg.Answer = append(msg.Answer, &dns.SOA{
			Hdr: dns.RR_Header{
				Name:   msg.Question[0].Name,
				Rrtype: dns.TypeSOA,
				Class:  dns.ClassINET,
				Ttl:    h.ttl,
			},
			Ns:      fmt.Sprintf("ns1.%s.", h.domain),
			Mbox:    fmt.Sprintf("hostmaster.%s.", h.domain),
			Serial:  h.soa,
			Refresh: 10800,
			Retry:   1800,
			Expire:  3600000,
			Minttl:  300,
		})

	case dns.TypeNS:
		msg.Answer = append(msg.Answer, &dns.NS{
			Hdr: dns.RR_Header{
				Name:   msg.Question[0].Name,
				Rrtype: dns.TypeNS,
				Class:  dns.ClassINET,
				Ttl:    h.ttl,
			},
			Ns: fmt.Sprintf("ns1.%s.", h.domain),
		})
		msg.Answer = append(msg.Answer, &dns.A{
			Hdr: dns.RR_Header{
				Name:   fmt.Sprintf("ns1.%s.", h.domain),
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    h.ttl,
			},
			A: h.ip,
		})
	}

	data, _ := msg.Pack()
	proto := fmt.Sprintf("DNS (%s)", dns.Type(r.Question[0].Qtype).String())

	if h.handlerFunc != nil {
		h.handlerFunc(w.RemoteAddr(), proto, data)
	}

	if err := w.WriteMsg(&msg); err != nil {
		log.Println(err)
	}
}

// Present allows Handler to satisfy challenge.Provider interface
func (h *handler) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	if values, ok := h.txtRecords.LoadOrStore(strings.ToLower(fqdn), []string{value}); ok {
		records := values.([]string)
		records = append(records, value)
		h.txtRecords.Store(strings.ToLower(fqdn), records)
	}

	// Increase SOA serial
	h.soa++

	return nil
}

// CleanUp allows Handler to satisfy challenge.Provider interface
func (h *handler) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	// Clear records
	h.txtRecords.Delete(strings.ToLower(fqdn))

	// Increase SOA serial
	h.soa++

	return nil
}
