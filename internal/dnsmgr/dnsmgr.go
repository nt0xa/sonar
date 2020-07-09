package dnsmgr

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils/tpl"
	"github.com/miekg/dns"
)

var recordsTpl = tpl.MustParse(`
{{- if .IP.To4 -}}
@ IN A    {{ .IP }}
* IN A    {{ .IP }}
@ IN AAAA ::ffff:{{ .IP }}
* IN AAAA ::ffff:{{ .IP }}
{{- else -}}
@ IN AAAA {{ .IP }}
* IN AAAA {{ .IP }}
{{- end }}
@ IN MX   10 mx
* IN MX   10 mx
`)

type DNSMgr struct {
	db           *database.DB
	origin       string
	dynamicRegex *regexp.Regexp

	static map[string][]dns.RR
}

func New(domain string, ip net.IP, subdomainPattern string, db *database.DB) (*DNSMgr, error) {

	re, err := regexp.Compile(fmt.Sprintf(`.*\.%s\.%s\.`,
		subdomainPattern, strings.ReplaceAll(domain, ".", "\\.")))

	if err != nil {
		return nil, fmt.Errorf("fail to compile payload regexp pattern: %w", err)
	}

	mgr := &DNSMgr{
		static:       make(map[string][]dns.RR),
		db:           db,
		dynamicRegex: re,
		origin:       domain,
	}

	data := struct {
		IP net.IP
	}{
		IP: ip,
	}

	records, err := tpl.RenderToString(recordsTpl, data)
	if err != nil {
		return nil, fmt.Errorf("fail to render default records template: %w", err)
	}

	rrs, err := parseZoneFile(strings.NewReader(records), domain)
	if err != nil {
		return nil, fmt.Errorf("fail to parse default records zone: %w", err)
	}

	for _, rr := range rrs {
		mgr.addStatic(rr)
	}

	return mgr, nil
}
