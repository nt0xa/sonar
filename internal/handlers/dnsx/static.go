package dnsx

import (
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"

	"github.com/bi-zone/sonar/internal/utils/tpl"
)

var recordsTpl = tpl.MustParse(`
@ IN 600 NS ns1
* IN 600 NS ns1
@ IN 600 NS ns2
* IN 600 NS ns2
{{ if .IP.To4 -}}
@ IN 600 A    {{ .IP }}
* IN 600 A    {{ .IP }}
@ IN AAAA ::ffff:{{ .IP }}
* IN AAAA ::ffff:{{ .IP }}
{{- else -}}
@ IN AAAA {{ .IP }}
* IN AAAA {{ .IP }}
{{- end }}
@ 600 IN MX   10 mx
* 600 IN MX   10 mx
`)

func (h *DNSX) addStaticRecords() error {
	data := struct {
		IP net.IP
	}{
		IP: h.ip,
	}

	records, err := tpl.RenderToString(recordsTpl, data)
	if err != nil {
		return fmt.Errorf("fail to render default records template: %w", err)
	}

	rrs, err := parseZoneFile(strings.NewReader(records), h.origin)
	if err != nil {
		return fmt.Errorf("fail to parse default records zone: %w", err)
	}

	for _, rr := range rrs {
		h.addStatic(rr)
	}

	return nil
}

func (h *DNSX) addStatic(rr dns.RR) {
	key := makeKey(rr.Header().Name, rr.Header().Rrtype)

	if _, ok := h.static[key]; !ok {
		h.static[key] = make([]dns.RR, 0)
	}

	h.static[key] = append(h.static[key], rr)
}

func (h *DNSX) delStatic(name string, qtype uint16) {
	key := makeKey(name, qtype)

	if _, ok := h.static[key]; !ok {
		return
	}

	delete(h.static, key)
}

func (h *DNSX) getStatic(name string, qtype uint16) []dns.RR {
	key := makeKey(name, qtype)

	if _, ok := h.static[key]; !ok {
		return nil
	}

	return h.static[key]
}
