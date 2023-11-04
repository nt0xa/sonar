package templates

import "fmt"

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

