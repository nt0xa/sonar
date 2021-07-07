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

type terminalHandler struct {
	domain string
}

func (h *terminalHandler) getDomain() string {
	return h.domain
}

func (h *terminalHandler) txtResult(txt string) {
	color.Println(txt)
}

func (h *terminalHandler) tplResult(tpl *template.Template, data interface{}) {
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

func (h *terminalHandler) UserCurrent(ctx context.Context, res actions.UserCurrentResult) {
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

func (h *terminalHandler) PayloadsCreate(ctx context.Context, res actions.PayloadsCreateResult) {
	h.tplResult(payloadTemplate, res)
}

func (h *terminalHandler) PayloadsList(ctx context.Context, res actions.PayloadsListResult) {
	h.tplResult(payloadsTemplate, res)
}

func (h *terminalHandler) PayloadsUpdate(ctx context.Context, res actions.PayloadsUpdateResult) {
	h.tplResult(payloadTemplate, res)
}

func (h *terminalHandler) PayloadsDelete(ctx context.Context, res actions.PayloadsDeleteResult) {
	h.txtResult(fmt.Sprintf("payload %q deleted", res.Name))
}

//
// DNS records
//

var (
	dnsRecord = `
{{- $r := . -}}
{{- range $value := .Values -}}
<bold>[{{ $r.Index }}]</> - {{ $r.Name }}.{{ $r.PayloadSubdomain }}.{{ domain }} {{ $r.TTL }} IN {{ $r.Type }} {{ $value }}
{{ end -}}`

	dnsRecordTemplate = tpl(dnsRecord)

	dnsRecordsTemplate = tpl(fmt.Sprintf(`
{{- range . -}}
%s
{{ else }}nothing found{{ end -}}`, dnsRecord))
)

func (h *terminalHandler) DNSRecordsCreate(ctx context.Context, res actions.DNSRecordsCreateResult) {
	h.tplResult(dnsRecordTemplate, res)
}

func (h *terminalHandler) DNSRecordsList(ctx context.Context, res actions.DNSRecordsListResult) {
	h.tplResult(dnsRecordsTemplate, res)
}

func (h *terminalHandler) DNSRecordsDelete(ctx context.Context, res actions.DNSRecordsDeleteResult) {
	h.txtResult("dns record deleted")
}

//
// HTTP routes
//

var (
	httpRoute = `
{{- $r := . -}}
<bold>[{{ $r.Index }}]</> - {{ $r.Method }} {{ $r.Path }} -> {{ $r.Code }}`

	httpRouteTemplate = tpl(httpRoute)

	httpRoutesTemplate = tpl(fmt.Sprintf(`
{{- range . -}}
%s
{{ else }}nothing found{{ end }}`, httpRoute))
)

func (h *terminalHandler) HTTPRoutesCreate(ctx context.Context, res actions.HTTPRoutesCreateResult) {
	h.tplResult(httpRouteTemplate, res)
}

func (h *terminalHandler) HTTPRoutesList(ctx context.Context, res actions.HTTPRoutesListResult) {
	h.tplResult(httpRoutesTemplate, res)
}

func (h *terminalHandler) HTTPRoutesDelete(ctx context.Context, res actions.HTTPRoutesDeleteResult) {
	h.txtResult("http route deleted")
}

//
// Users
//

func (h *terminalHandler) UsersCreate(ctx context.Context, res actions.UsersCreateResult) {
	h.txtResult(fmt.Sprintf("user %q created", res.Name))
}

func (h *terminalHandler) UsersDelete(ctx context.Context, res actions.UsersDeleteResult) {
	h.txtResult(fmt.Sprintf("user %q deleted", res.Name))
}

//
// Events
//

var (
	event = `
{{- $e := . -}}
<bold>[{{ $e.Index }}]</> - {{ $e.Protocol | upper }} from {{ $e.RemoteAddr }} at {{ $e.ReceivedAt }}`

	eventTemplate = tpl(event + `

{{ $e.RW | b64dec }}
`)

	eventsTemplate = tpl(fmt.Sprintf(`
{{- range . -}}
%s
{{ else }}nothing found{{ end -}}`, event))
)

func (h *terminalHandler) EventsList(ctx context.Context, res actions.EventsListResult) {
	h.tplResult(eventsTemplate, res)
}

func (h *terminalHandler) EventsGet(ctx context.Context, res actions.EventsGetResult) {
	h.tplResult(eventTemplate, res)
}
