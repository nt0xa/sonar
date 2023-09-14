package results

import (
	"fmt"
	"io"
	"strings"

	"html/template"
	htmltemplate "html/template"
	texttemplate "text/template"

	"github.com/Masterminds/sprig/v3"
)

type Template interface {
	Execute(io.Writer, any) error
}

func MakeTemplate(s string, opts TemplateOptions) Template {

	// Replace all pseudotags like "<bold>" or "<code>" using
	// provided markup map.
	// TODO: maybe move to Execute function, after executing template
	for tag, replacement := range opts.Markup {
		s = strings.ReplaceAll(s, tag, replacement)
	}

	if opts.HTML {
		return htmltemplate.Must(htmltemplate.
			New("").
			Funcs(sprig.FuncMap()).
			Funcs(opts.ExtraFuncs).
			Parse(s),
		)
	} else {
		return texttemplate.Must(texttemplate.
			New("").
			Funcs(sprig.FuncMap()).
			Funcs(opts.ExtraFuncs).
			Parse(s),
		)
	}

}

type TemplateOptions struct {
	Markup     map[string]string
	ExtraFuncs template.FuncMap
	HTML       bool
}

func DefaultTemplates(opts TemplateOptions) map[string]Template {
	return map[string]Template{
		"profile/get":        MakeTemplate(profileGet, opts),
		"payloads/get":       MakeTemplate(payloadsGet, opts),
		"payloads/list":      MakeTemplate(payloadsList, opts),
		"payloads/create":    MakeTemplate(payloadsCreate, opts),
		"payloads/update":    MakeTemplate(payloadsUpdate, opts),
		"payloads/delete":    MakeTemplate(payloadsDelete, opts),
		"dns-records/list":   MakeTemplate(dnsRecordsList, opts),
		"dns-records/create": MakeTemplate(dnsRecordsCreate, opts),
		"dns-records/delete": MakeTemplate(dnsRecordsDelete, opts),
		"http-routes/list":   MakeTemplate(httpRoutesList, opts),
		"http-routes/create": MakeTemplate(httpRoutesCreate, opts),
		"http-routes/delete": MakeTemplate(httpRoutesDelete, opts),
		"events/list":        MakeTemplate(eventsList, opts),
		"events/get":         MakeTemplate(eventsGet, opts),
		"text":               MakeTemplate(text, opts),
		"error":              MakeTemplate(err, opts),
	}
}

//
// Profile
//

var profileGet = `
<bold>Name:</bold> <code>{{ .Name }}</code>
{{ if .Params.TelegramID }}<bold>Telegram ID:</bold> <code>{{ .Params.TelegramID }}</code>
{{ end -}}
{{ if .Params.LarkUserID }}<bold>Lark ID:</bold> <code>{{ .Params.LarkUserID }}</code>
{{ end -}}
<bold>API token:</bold> <code>{{ .Params.APIToken }}</code>
<bold>Admin:</bold> <code>{{ .IsAdmin }}</code>`

//
// Payloads
//

var payload = `
{{- $p := . -}}
<bold>[{{ $p.Name }}]</bold> - <code>{{ $p.Subdomain }}.{{ domain }}</code> ({{ $p.NotifyProtocols | join ", " }}) ({{ $p.StoreEvents }})`

var payloadsGet = payload
var payloadsList = fmt.Sprintf(`{{ range . }}%s
{{ else }}nothing found{{ end }}`, payload)
var payloadsCreate = payload
var payloadsUpdate = payload
var payloadsDelete = `payload "{{ .Name }}" deleted`

//
// DNS records
//

var dnsRecord = `
{{- $r := . -}}
{{- range $value := .Values -}}
<bold>[{{ $r.Index }}]</bold> - {{ $r.Name }}.{{ $r.PayloadSubdomain }}.{{ domain }} {{ $r.TTL }} IN {{ $r.Type }} {{ $value }}
{{ end -}}`

var dnsRecordsList = fmt.Sprintf(`
{{- range . -}}
%s
{{ else }}nothing found{{ end }}`, dnsRecord)
var dnsRecordsCreate = dnsRecord
var dnsRecordsDelete = `dns record deleted`

//
// HTTP routes
//

var httpRoute = `
{{- $r := . -}}
<bold>[{{ $r.Index }}]</bold> - {{ $r.Method }} {{ $r.Path }} -> {{ $r.Code }}`

var httpRoutesList = fmt.Sprintf(`
{{- range . -}}
%s
{{ else }}nothing found{{ end }}`, httpRoute)
var httpRoutesCreate = httpRoute
var httpRoutesDelete = "http route deleted"

//
// Users
//

var usersCreate = `user "{{ .Name }}" created`
var usersDelete = `user "{{ .Name }}" deleted`

//
// Events
//

var event = `
{{- $e := . -}}
<bold>[{{ $e.Index }}]</bold> - {{ $e.Protocol | upper }} from {{ $e.RemoteAddr }} at {{ $e.ReceivedAt }}`

var eventsGet = event + `

<pre>{{ $e.RW | b64dec }}</pre>`

var eventsList = fmt.Sprintf(`
{{- range . -}}
%s
{{ else }}nothing found{{ end -}}`, event)

//
// Text
//

var text = "{{ .Text }}"

//
// Error
//

var err = "<error>{{ .Error }}</error>"
