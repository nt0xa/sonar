package templates

import (
	"fmt"

	"github.com/nt0xa/sonar/internal/actions"
)

var templatesMap = map[string]string{
	NotificationHeaderID: notificationHeader,
	NotificationBodyID:   notificationBody,

	actions.ProfileGetResultID: profileGet,

	actions.PayloadsListResultID:   payloadsList,
	actions.PayloadsCreateResultID: payloadsCreate,
	actions.PayloadsUpdateResultID: payloadsUpdate,
	actions.PayloadsDeleteResultID: payloadsDelete,
	actions.PayloadsClearResultID:  payloadsClear,

	actions.DNSRecordsListResultID:   dnsRecordsList,
	actions.DNSRecordsCreateResultID: dnsRecordsCreate,
	actions.DNSRecordsDeleteResultID: dnsRecordsDelete,
	actions.DNSRecordsClearResultID:  dnsRecordsClear,

	actions.HTTPRoutesListResultID:   httpRoutesList,
	actions.HTTPRoutesCreateResultID: httpRoutesCreate,
	actions.HTTPRoutesUpdateResultID: httpRoutesUpdate,
	actions.HTTPRoutesDeleteResultID: httpRoutesDelete,
	actions.HTTPRoutesClearResultID:  httpRoutesClear,

	actions.EventsListResultID: eventsList,
	actions.EventsGetResultID:  eventsGet,
}

//
// Profile
//

var profileGet = `<bold>Name:</bold> <code>{{ .Name }}</code>
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

var payloadsList = fmt.Sprintf(`{{ range . }}%s
{{ else }}nothing found{{ end }}`, payload)
var payloadsCreate = payload
var payloadsUpdate = payload
var payloadsDelete = `payload "{{ .Name }}" deleted`
var payloadsClear = `{{ len . }} payloads deleted`

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
var dnsRecordsClear = `{{ len . }} dns records deleted`

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
var httpRoutesUpdate = httpRoute
var httpRoutesDelete = "http route deleted"
var httpRoutesClear = `{{ len . }} http routes deleted`

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
<bold>[{{ $e.Index }}]</bold> - {{ $e.Protocol | upper }} from {{ $e.RemoteAddr }} {{ $e.ReceivedAt.Format "on 02 Jan 2006 at 15:04:05 MST" }}`

var eventsGet = event + `
{{- $protocol := $e.Protocol }}

<pre>
{{ $e.RW | b64dec }}
</pre>`

var eventsList = fmt.Sprintf(`
{{- range . -}}
%s
{{ else }}nothing found{{ end -}}`, event)

//
// Notification
//

const (
	NotificationHeaderID = "notification/header"
	NotificationBodyID   = "notification/body"
)

var notificationHeader = `
{{- $category := .Event.Protocol.Category.String -}}
#{{ .Payload.Name }} {{ if eq (upper $category) "FTP" -}}
📂
{{- else if eq (upper $category) "SMTP" -}}
📧
{{- else if eq (upper $category) "DNS" -}}
🔎
{{- else if eq (upper $category) "HTTP" -}}
🌐
{{- else -}}
❓
{{- end }} <bold>{{ .Event.Protocol.String | upper }}</bold>`

var notificationBody = `
{{- $protocol := .Event.Protocol.String -}}
📡 <bold>IP:</bold> <code>{{ regexReplaceAll ":[0-9]+$" .Event.RemoteAddr "" }}</code>
📆 <bold>Time:</bold> {{ .Event.ReceivedAt.Format "02 Jan 2006 15:04:05 MST" }}
{{- $geoip := .Event.Meta.geoip }}
{{- if $geoip }}
📍 <bold>Location:</bold> {{ $geoip.country.flagEmoji }} {{ $geoip.country.name }}, {{ $geoip.city }}
{{- if $geoip.asn }}
🏢 <bold>Org:</bold> {{ $geoip.asn.org }} (AS{{ $geoip.asn.number }})
{{- end }}
{{- end }}
{{- $email := .Event.Meta.email }}
{{- range $from := $email.from }}
👤 <bold>From:</bold> <code>{{ $from.email }}</code>
{{- end }}
{{- if $email.subject }}
💬 <bold>Subject:</bold> {{ $email.subject }}
{{- end }}

<pre>
{{ if $email.text -}}
{{ $email.text }}
{{- else -}}
{{ printf "%s" .Event.RW }}
{{- end -}}
</pre>`
