package server

import (
	"net"
	"strings"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/dnsdb"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/tpl"
	"github.com/bi-zone/sonar/pkg/dnsrec"
	"github.com/bi-zone/sonar/pkg/dnsutils"
	"github.com/bi-zone/sonar/pkg/dnsx"
	"github.com/fatih/structs"
	"github.com/miekg/dns"
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
`)

func DNSDefaultRecords(origin string, ip net.IP) *dnsrec.Records {
	s, _ := tpl.RenderToString(dnsTemplate, ip)
	rrs := dnsutils.Must(dnsutils.ParseRecords(s, origin))
	return dnsrec.New(rrs)
}

func DNSHandler(db *database.DB, origin string, ip net.IP, notify func(*dnsx.Event)) dnsx.HandlerProvider {
	// Do not handle DNS queries which are not subdomains of the origin.
	h := dns.NewServeMux()

	h.Handle(origin,
		dnsx.NotifyHandler(
			notify,
			dnsx.ChainHandler(
				// Database records.
				&dnsdb.Records{DB: db, Origin: origin},
				// Fallback records.
				dnsx.RecordSetHandler(DNSDefaultRecords(origin, ip)),
			),
		),
	)

	return dnsx.ChallengeHandler(h)
}

func DNSEvent(e *dnsx.Event) *models.Event {

	type Question struct {
		Name  string `structs:"name"`
		Qtype string `structs:"qtype"`
	}

	type Answer struct {
		Name  string `structs:"name"`
		Rtype string `structs:"rtype"`
		TTL   uint32 `structs:"ttl"`
	}

	type Meta struct {
		Question Question `structs:"question"`
		Answer   []Answer `structs:"answer"`
	}

	meta := new(Meta)
	w := ""

	meta.Question.Name = strings.Trim(e.Msg.Question[0].Name, ".")
	meta.Question.Qtype = dnsutils.QtypeString(e.Msg.Question[0].Qtype)

	if len(e.Msg.Answer) > 0 {
		for _, rr := range e.Msg.Answer {
			meta.Answer = append(meta.Answer, Answer{
				Name:  strings.Trim(rr.Header().Name, "."),
				Rtype: dnsutils.QtypeString(rr.Header().Rrtype),
				TTL:   rr.Header().Ttl,
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
