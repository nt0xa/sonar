package dns

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/fatih/structs"
	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/miekg/dns"
)

type Server struct {
	addr    string
	ip      net.IP
	domain  string
	options *options

	soa        uint32
	txtRecords sync.Map
}

type Meta struct {
	Qtype string
}

func New(addr, domain string, ip net.IP, opts ...Option) *Server {
	options := defaultOptions

	for _, fopt := range opts {
		fopt(&options)
	}

	return &Server{
		options: &options,
		addr:    addr,
		ip:      ip,
		domain:  domain,
	}
}

func (s *Server) SetOption(opt Option) {
	opt(s.options)
}

func (s *Server) ListenAndServe() error {
	srv := &dns.Server{
		Addr:              s.addr,
		Net:               "udp",
		Handler:           s,
		NotifyStartedFunc: s.options.notifyStartedFunc,
	}

	return srv.ListenAndServe()
}

// ServeDNS allows handler to satisfy dns.Handler interface
func (s *Server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}

	msg.SetReply(r)

	switch r.Question[0].Qtype {

	case dns.TypeAAAA:
		msg.Answer = append(msg.Answer, &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   msg.Question[0].Name,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    s.options.ttl,
			},
			AAAA: s.ip,
		})

	case dns.TypeA:
		msg.Answer = append(msg.Answer, &dns.A{
			Hdr: dns.RR_Header{
				Name:   msg.Question[0].Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    s.options.ttl,
			},
			A: s.ip,
		})

	case dns.TypeMX:
		msg.Answer = append(msg.Answer, &dns.MX{
			Hdr: dns.RR_Header{
				Name:   msg.Question[0].Name,
				Rrtype: dns.TypeMX,
				Class:  dns.ClassINET,
				Ttl:    s.options.ttl,
			},
			Preference: 10,
			Mx:         fmt.Sprintf("mx.%s.", s.domain),
		})

	case dns.TypeTXT:
		if values, ok := s.txtRecords.Load(strings.ToLower(msg.Question[0].Name)); ok {
			records := values.([]string)
			for _, value := range records {
				msg.Answer = append(msg.Answer, &dns.TXT{
					Hdr: dns.RR_Header{
						Name:   msg.Question[0].Name,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    s.options.ttl,
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
				Ttl:    s.options.ttl,
			},
			Ns:      fmt.Sprintf("ns1.%s.", s.domain),
			Mbox:    fmt.Sprintf("hostmaster.%s.", s.domain),
			Serial:  s.soa,
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
				Ttl:    s.options.ttl,
			},
			Ns: fmt.Sprintf("ns1.%s.", s.domain),
		})
		msg.Answer = append(msg.Answer, &dns.A{
			Hdr: dns.RR_Header{
				Name:   fmt.Sprintf("ns1.%s.", s.domain),
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    s.options.ttl,
			},
			A: s.ip,
		})
	}

	data, _ := msg.Pack()
	meta := Meta{
		Qtype: dns.Type(r.Question[0].Qtype).String(),
	}

	s.options.notifyRequestFunc(w.RemoteAddr(), data, structs.Map(meta))

	if err := w.WriteMsg(&msg); err != nil {
		log.Println(err)
	}
}

// Present allows Server to satisfy challenge.Provider interface
func (s *Server) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	if values, ok := s.txtRecords.LoadOrStore(strings.ToLower(fqdn), []string{value}); ok {
		records := values.([]string)
		records = append(records, value)
		s.txtRecords.Store(strings.ToLower(fqdn), records)
	}

	// Increase SOA serial
	s.soa++

	return nil
}

// CleanUp allows Server to satisfy challenge.Provider interface
func (s *Server) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	// Clear records
	s.txtRecords.Delete(strings.ToLower(fqdn))

	// Increase SOA serial
	s.soa++

	return nil
}
