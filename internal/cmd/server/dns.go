package server

import (
	"io/ioutil"
	"net"
	"strings"

	"github.com/fatih/structs"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/miekg/dns"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/dnsdb"
	"github.com/nt0xa/sonar/internal/utils/tpl"
	"github.com/nt0xa/sonar/pkg/dnsx"
)

var dnsTemplate = tpl.MustParse(`
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
@ SOA ns1 hostmaster 1337 86400 7200 4000000 11200
* SOA ns1 hostmaster 1337 86400 7200 4000000 11200
`)

type DNSConfig struct {
	Zone string `json:"zone"`
}

func (c DNSConfig) Validate() error {
	return validation.ValidateStruct(&c)
}

func parseDNSRecords(s, origin string, ip net.IP) *dnsx.Records {
	rrs, err := dnsx.ParseRecords(s, origin)
	if err != nil {
		panic(err)
	}
	return dnsx.NewRecords(rrs)
}

func DNSDefaultRecords(origin string, ip net.IP) *dnsx.Records {
	s, _ := tpl.RenderToString(dnsTemplate, ip)
	return parseDNSRecords(s, origin, ip)
}

func DNSZoneFileRecords(filePath, origin string, ip net.IP) *dnsx.Records {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil
	}
	return parseDNSRecords(string(data), origin, ip)
}

func DNSHandler(cfg *DNSConfig, db *database.DB, origin string, ip net.IP, notify func(*dnsx.Event)) dnsx.HandlerProvider {
	// Do not handle DNS queries which are not subdomains of the origin.
	h := dns.NewServeMux()

	var fallback dns.Handler

	defaultRecords := DNSDefaultRecords(origin, ip)

	if extraRecords := DNSZoneFileRecords(cfg.Zone, origin, ip); extraRecords != nil {
		fallback = dnsx.ChainHandler(extraRecords, dnsx.RecordSetHandler(defaultRecords))
	} else {
		fallback = dnsx.RecordSetHandler(defaultRecords)
	}

	h.Handle(origin,
		dnsx.NotifyHandler(
			notify,
			dnsx.ChainHandler(
				// Database records.
				&dnsdb.Records{DB: db, Origin: origin},
				// Fallback records.
				fallback,
			),
		),
	)

	return dnsx.ChallengeHandler(h)
}

func DNSEvent(e *dnsx.Event) *models.Event {

	type Question struct {
		Name string `structs:"name"`
		Type string `structs:"type"`
	}

	type Answer struct {
		Name string `structs:"name"`
		Type string `structs:"type"`
		TTL  uint32 `structs:"ttl"`
	}

	type Meta struct {
		Question Question `structs:"question"`
		Answer   []Answer `structs:"answer"`
	}

	meta := new(Meta)
	w := ""

	meta.Question.Name = strings.Trim(e.Msg.Question[0].Name, ".")
	meta.Question.Type = dnsx.QtypeString(e.Msg.Question[0].Qtype)

	if len(e.Msg.Answer) > 0 {
		for _, rr := range e.Msg.Answer {
			meta.Answer = append(meta.Answer, Answer{
				Name: strings.Trim(rr.Header().Name, "."),
				Type: dnsx.QtypeString(rr.Header().Rrtype),
				TTL:  rr.Header().Ttl,
			})
		}
		w = e.Msg.Answer[0].String()
	}

	return &models.Event{
		Protocol:   models.ProtoDNS,
		R:          []byte(e.Msg.Question[0].String()),
		W:          []byte(w),
		RW:         []byte(e.Msg.String()),
		RemoteAddr: e.RemoteAddr.String(),
		ReceivedAt: e.ReceivedAt,
		Meta:       models.Meta(structs.Map(meta)),
	}
}
