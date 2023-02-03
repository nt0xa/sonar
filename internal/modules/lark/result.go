package lark

import (
	"context"
	"fmt"

	"github.com/russtone/sonar/internal/actions"
)

// Ensure Lark implemenents actions.ResultHandler interface.
var _ actions.ResultHandler = (*Lark)(nil)

// TODO: move into common module
var (
	helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}
{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

	usageTemplate = `
Usage:{{if .Runnable}}{{if .HasParent}}
  {{.UseLine | replace "sonar " "/"}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
	{{if .HasParent}}{{.CommandPath | replace "sonar " "/"}} {{else}}/{{end}}[command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  /{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{if .HasParent}}{{.CommandPath | replace "sonar " "/"}} {{else}}/{{end}}[command] --help" for more information about a command.{{end}}
`

	codeTemplate = tpl(`{{ . }}`)
)

//
// User
//

var userCurrentTemplate = tpl("" +
	"**Lark ID:** {{ .Params.LarkUserID }}\n" +
	"**API token:** {{ .Params.APIToken }}",
)

func (lrk *Lark) UserCurrent(ctx context.Context, res actions.UserCurrentResult) {
	lrk.tplResult(ctx, userCurrentTemplate, res)
}

//
// Payloads
//

var (
	payload = `**[{{ .Name }}]** - {{ .Subdomain }}.{{ domain }} ({{ .NotifyProtocols | join ", " }}) ({{ .StoreEvents }})`

	payloadTemplate = tpl(payload)

	payloadsTemplate = tpl(fmt.Sprintf(`{{ range . }}%s
{{ else }}nothing found{{ end }}`, payload))
)

func (lrk *Lark) PayloadsCreate(ctx context.Context, res actions.PayloadsCreateResult) {
	lrk.tplResult(ctx, payloadTemplate, res)
}

func (lrk *Lark) PayloadsList(ctx context.Context, res actions.PayloadsListResult) {
	lrk.tplResult(ctx, payloadsTemplate, res)
}

func (lrk *Lark) PayloadsUpdate(ctx context.Context, res actions.PayloadsUpdateResult) {
	lrk.tplResult(ctx, payloadTemplate, res)
}

func (lrk *Lark) PayloadsDelete(ctx context.Context, res actions.PayloadsDeleteResult) {
	lrk.txtResult(ctx, fmt.Sprintf("payload %q deleted", res.Name))
}

//
// DNS records
//

var (
	dnsRecord = `
{{- $r := . -}}
{{- range $value := .Values -}}
**[{{ $r.Index }}]** - {{ $r.Name }}.{{ $r.PayloadSubdomain }}.{{ domain }} {{ $r.TTL }} IN {{ $r.Type }} {{ $value }}
{{ end -}}`

	dnsRecordTemplate = tpl(dnsRecord)

	dnsRecordsTemplate = tpl(fmt.Sprintf(`
{{- range . -}}
%s
{{ else }}nothing found{{ end }}`, dnsRecord))
)

func (lrk *Lark) DNSRecordsCreate(ctx context.Context, res actions.DNSRecordsCreateResult) {
	lrk.tplResult(ctx, dnsRecordTemplate, res)
}

func (lrk *Lark) DNSRecordsList(ctx context.Context, res actions.DNSRecordsListResult) {
	lrk.tplResult(ctx, dnsRecordsTemplate, res)
}

func (lrk *Lark) DNSRecordsDelete(ctx context.Context, res actions.DNSRecordsDeleteResult) {
	lrk.txtResult(ctx, "dns record deleted")
}

//
// HTTP routes
//

var (
	httpRoute = `
{{- $r := . -}}
**[{{ $r.Index }}] - **{{ $r.Method }} {{ $r.Path }} -> {{ $r.Code }}`

	httpRouteTemplate = tpl(httpRoute)

	httpRoutesTemplate = tpl(fmt.Sprintf(`
{{- range . -}}
%s
{{ else }}nothing found{{ end }}`, httpRoute))
)

func (lrk *Lark) HTTPRoutesCreate(ctx context.Context, res actions.HTTPRoutesCreateResult) {
	lrk.tplResult(ctx, httpRouteTemplate, res)
}

func (lrk *Lark) HTTPRoutesList(ctx context.Context, res actions.HTTPRoutesListResult) {
	lrk.tplResult(ctx, httpRoutesTemplate, res)
}

func (lrk *Lark) HTTPRoutesDelete(ctx context.Context, res actions.HTTPRoutesDeleteResult) {
	lrk.txtResult(ctx, "http route deleted")
}

//
// Users
//

func (lrk *Lark) UsersCreate(ctx context.Context, res actions.UsersCreateResult) {
	lrk.txtResult(ctx, fmt.Sprintf("user %q created", res.Name))
}

func (lrk *Lark) UsersDelete(ctx context.Context, res actions.UsersDeleteResult) {
	lrk.txtResult(ctx, fmt.Sprintf("user %q deleted", res.Name))
}

//
// Events
//

var (
	eventCommon = `
{{- $e := . -}}
**[{{ $e.Index }}]** - {{ $e.Protocol | upper }} from {{ $e.RemoteAddr }} at {{ $e.ReceivedAt }}`

	eventTemplate = tpl(eventCommon + `

{{ $e.RW | b64dec }}
`)

	eventsTemplate = tpl(fmt.Sprintf(`
{{- range . -}}
%s
{{ else }}nothing found{{ end -}}`, eventCommon))
)

func (lrk *Lark) EventsList(ctx context.Context, res actions.EventsListResult) {
	lrk.tplResult(ctx, eventsTemplate, res)
}

func (lrk *Lark) EventsGet(ctx context.Context, res actions.EventsGetResult) {
}
