package dnsdef

import (
	"net"

	"github.com/bi-zone/sonar/internal/protocols/dnsx/dnsrec"
	"github.com/bi-zone/sonar/internal/protocols/dnsx/dnsutils"
	"github.com/bi-zone/sonar/internal/utils/tpl"
)

var records = tpl.MustParse(`
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

// Records returns default DNS records.
func Records(origin string, ip net.IP) (*dnsrec.Records, error) {
	s, err := tpl.RenderToString(records, ip)
	if err != nil {
		return nil, err
	}

	rrs, err := dnsutils.ParseRecords(s, origin)
	if err != nil {
		return nil, err
	}

	return dnsrec.New(rrs), nil
}
