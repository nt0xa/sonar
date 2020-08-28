package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/bi-zone/sonar/internal/actions"
	"github.com/gookit/color"
)

func tpl(s string) *template.Template {
	return template.Must(template.
		New("").
		Funcs(sprig.TxtFuncMap()).
		Funcs(template.FuncMap{
			// This is nesessary for templates to compile.
			// It will be replaced later with correct function.
			"domain": func() string { return "" },
		}).
		Parse(s),
	)
}

type handler struct {
	domain string
}

func (h *handler) getDomain() string {
	return h.domain
}

func (h *handler) txtResult(txt string) {
	color.Println(txt)
}

func (h *handler) tplResult(tpl *template.Template, data interface{}) {
	buf := &bytes.Buffer{}

	tpl.Funcs(template.FuncMap{
		"domain": h.getDomain,
	})

	if err := tpl.Execute(buf, data); err != nil {
		color.Error.Println(err)
		os.Exit(1)
	}

	color.Println(strings.TrimRight(buf.String(), "\n"))
}

//
// User
//

var userCurrentTemplate = tpl("" +
	"<bold>Telegram ID:</> {{ .TelegramID }}\n" +
	"<bold>API token:</> {{ .APIToken }}",
)

func (h *handler) UserCurrent(ctx context.Context, res actions.UserCurrentResult) {
	h.tplResult(userCurrentTemplate, res)
}

//
// Payloads
//

var (
	payload = `<bold>[{{ .Name }}]</> - {{ .Subdomain }}.{{ domain }} ({{ .NotifyProtocols | join ", " }})`

	payloadTemplate = tpl(payload)

	payloadsTemplate = tpl(fmt.Sprintf(`{{ range . }}%s
{{ else }}nothing found{{ end }}`, payload))
)

func (h *handler) PayloadsCreate(ctx context.Context, res actions.PayloadsCreateResult) {
	h.tplResult(payloadTemplate, res)
}

func (h *handler) PayloadsList(ctx context.Context, res actions.PayloadsListResult) {
	h.tplResult(payloadsTemplate, res)
}

func (h *handler) PayloadsUpdate(ctx context.Context, res actions.PayloadsUpdateResult) {
	h.tplResult(payloadTemplate, res)
}

func (h *handler) PayloadsDelete(ctx context.Context, res actions.PayloadsDeleteResult) {
	h.txtResult(fmt.Sprintf("payload %q deleted", res.Name))
}

//
// DNS records
//

var (
	dnsRecord = `
{{- $p := $.Payload -}}
{{- range $value := $r.Values -}}
<bold>{{ $r.Name }}.{{ $p.Subdomain }}.{{ domain }}</> {{ $r.TTL }} IN {{ $r.Type }} {{ $value }}
{{ end -}}`

	dnsRecordTemplate = tpl(`{{ $r := .Record }}` + dnsRecord)

	dnsRecordsTemplate = tpl(fmt.Sprintf(`
{{- range .Records -}}
{{ $r := . }}
%s
{{ else }}nothing found{{ end -}}`, dnsRecord))
)

func (h *handler) DNSRecordsCreate(ctx context.Context, res actions.DNSRecordsCreateResult) {
	h.tplResult(dnsRecordTemplate, res)
}

func (h *handler) DNSRecordsList(ctx context.Context, res actions.DNSRecordsListResult) {
	h.tplResult(dnsRecordsTemplate, res)
}

func (h *handler) DNSRecordsDelete(ctx context.Context, res actions.DNSRecordsDeleteResult) {
	h.txtResult("dns record deleted")
}

//
// Users
//

func (h *handler) UsersCreate(ctx context.Context, res actions.UsersCreateResult) {
	h.txtResult(fmt.Sprintf("user %q created", res.Name))
}

func (h *handler) UsersDelete(ctx context.Context, res actions.UsersDeleteResult) {
	h.txtResult(fmt.Sprintf("user %q deleted", res.Name))
}
