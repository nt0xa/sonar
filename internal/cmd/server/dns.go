package server

import (
	"context"
	"net"
	"os"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/miekg/dns"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/dnsdb"
	"github.com/nt0xa/sonar/internal/utils/tpl"
	"github.com/nt0xa/sonar/pkg/dnsx"
	"github.com/nt0xa/sonar/pkg/telemetry"
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

@ 60 IN CAA 0 issue "letsencrypt.org"
@ SOA ns1 hostmaster 1337 86400 7200 4000000 11200
* SOA ns1 hostmaster 1337 86400 7200 4000000 11200
`)

func parseDNSRecords(s, origin string, _ net.IP) *dnsx.Records {
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
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	return parseDNSRecords(string(data), origin, ip)
}

func DNSHandler(
	cfg *DNSConfig,
	db *database.DB,
	tel telemetry.Telemetry,
	origin string,
	ip net.IP,
	notify func(context.Context, *dnsx.Event),
) dnsx.HandlerProvider {
	// Do not handle DNS queries which are not subdomains of the origin.
	h := dnsx.NewServeMux()

	var fallback dnsx.Handler

	defaultRecords := DNSDefaultRecords(origin, ip)

	if extraRecords := DNSZoneFileRecords(cfg.Zone, origin, ip); extraRecords != nil {
		fallback = dnsx.ChainHandler(extraRecords, dnsx.RecordSetHandler(defaultRecords))
	} else {
		fallback = dnsx.RecordSetHandler(defaultRecords)
	}

	h.Handle(origin,
		DNSTelemetryHandler(
			tel,
			dnsx.NotifyHandler(
				notify,
				dnsx.ChainHandler(
					// Database records.
					&dnsdb.Records{DB: db, Origin: origin},
					// Fallback records.
					fallback,
				),
			),
		),
	)

	return dnsx.ChallengeHandler(h)
}

func DNSTelemetryHandler(tel telemetry.Telemetry, next dnsx.Handler) dnsx.Handler {
	queryDuration, err := tel.NewInt64Histogram(
		"dns.query.duration",
		"ms",
		"DNS query duration",
	)
	if err != nil {
		panic(err)
	}

	counter, err := tel.NewInt64UpDownCounter(
		"dns.queries.inflight",
		"{count}",
		"Number of queries currently being processed by the server",
	)
	if err != nil {
		panic(err)
	}

	return dnsx.HandlerFunc(func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) {
		start := time.Now()
		ctx, id := withEventID(ctx)

		ctx, span := tel.TraceStart(ctx, "dns", trace.WithAttributes(
			attribute.String("event.id", id.String()),
			attribute.String("dns.query.name", r.Question[0].Name),
			attribute.String("dns.query.type", dnsx.QtypeString(r.Question[0].Qtype)),
		))
		defer span.End()

		counter.Add(ctx, 1)

		next.ServeDNS(ctx, w, r)

		counter.Add(ctx, -1)
		queryDuration.Record(ctx, time.Since(start).Milliseconds())
	})
}

func DNSEvent(e *dnsx.Event) *models.Event {
	w := ""

	meta := models.Meta{
		DNS: &models.DNSMeta{
			Question: &models.DNSQuestion{
				Name: strings.Trim(e.Msg.Question[0].Name, "."),
				Type: dnsx.QtypeString(e.Msg.Question[0].Qtype),
			},
		},
	}

	if len(e.Msg.Answer) > 0 {
		meta.DNS.Answer = make([]models.DNSAnswer, 0, len(e.Msg.Answer))
		for _, rr := range e.Msg.Answer {
			meta.DNS.Answer = append(meta.DNS.Answer, models.DNSAnswer{
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
		Meta:       meta,
	}
}
