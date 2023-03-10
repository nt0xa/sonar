package telegram

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/Masterminds/sprig"
	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/actionsdb"
	"github.com/russtone/sonar/internal/utils/errors"
)

var (
	helpTemplate = `<code>{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}
{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}</code>`

	usageTemplate = `<code>
Usage:{{if .Runnable}}{{if .HasParent}}
  {{.UseLine | replace "sonarctl " "/"}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
	{{if .HasParent}}{{.CommandPath | replace "sonarctl " "/"}} {{else}}/{{end}}[command]{{end}}{{if gt (len .Aliases) 0}}

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

Use "{{if .HasParent}}{{.CommandPath | replace "sonarctl " "/"}} {{else}}/{{end}}[command] --help" for more information about a command.{{end}}
</code>`

	codeTemplate = tpl(`<code>{{ . }}</code>`)
)

func tpl(s string) *template.Template {
	return template.Must(template.
		New("").
		Funcs(sprig.FuncMap()).
		Funcs(template.FuncMap{
			// This is nesessary for templates to compile.
			// It will be replaced later with correct function.
			"domain": func() string { return "" },
		}).
		Parse(s),
	)
}

func (tg *Telegram) getDomain() string {
	return tg.domain
}

func (tg *Telegram) txtResult(ctx context.Context, txt string) {
	u, err := actionsdb.GetUser(ctx)
	if err != nil {
		return
	}

	tg.txtMessage(u.Params.TelegramID, txt)
}

func (tg *Telegram) tplResult(ctx context.Context, tpl *template.Template, data interface{}) {
	u, err := actionsdb.GetUser(ctx)
	if err != nil {
		return
	}

	tpl.Funcs(template.FuncMap{
		"domain": tg.getDomain,
	})

	buf := &bytes.Buffer{}

	if err := tpl.Execute(buf, data); err != nil {
		tg.handleError(u.Params.TelegramID, errors.Internal(err))
	}

	tg.htmlMessage(u.Params.TelegramID, buf.String())
}

//
// User
//

var profileGetTemplate = tpl("" +
	"<b>Telegram ID:</b> <code>{{ .Params.TelegramID }}</code>\n" +
	"<b>API token:</b> <code>{{ .Params.APIToken }}</code>",
)

func (tg *Telegram) ProfileGet(ctx context.Context, res actions.ProfileGetResult) {
	tg.tplResult(ctx, profileGetTemplate, res)
}

//
// Payloads
//

var (
	payload = `<b>[{{ .Name }}]</b> - <code>{{ .Subdomain }}.{{ domain }}</code> ({{ .NotifyProtocols | join ", " }}) ({{ .StoreEvents }})`

	payloadTemplate = tpl(payload)

	payloadsTemplate = tpl(fmt.Sprintf(`{{ range . }}%s
{{ else }}nothing found{{ end }}`, payload))
)

func (tg *Telegram) PayloadsCreate(ctx context.Context, res actions.PayloadsCreateResult) {
	tg.tplResult(ctx, payloadTemplate, res)
}

func (tg *Telegram) PayloadsList(ctx context.Context, res actions.PayloadsListResult) {
	tg.tplResult(ctx, payloadsTemplate, res)
}

func (tg *Telegram) PayloadsUpdate(ctx context.Context, res actions.PayloadsUpdateResult) {
	tg.tplResult(ctx, payloadTemplate, res)
}

func (tg *Telegram) PayloadsDelete(ctx context.Context, res actions.PayloadsDeleteResult) {
	tg.txtResult(ctx, fmt.Sprintf("payload %q deleted", res.Name))
}

//
// DNS records
//

var (
	dnsRecord = `
{{- $r := . -}}
{{- range $value := .Values -}}
<b>[{{ $r.Index }}] - </b><code>{{ $r.Name }}.{{ $r.PayloadSubdomain }}.{{ domain }}</code><code> {{ $r.TTL }} IN {{ $r.Type }} {{ $value }}</code>
{{ end -}}`

	dnsRecordTemplate = tpl(dnsRecord)

	dnsRecordsTemplate = tpl(fmt.Sprintf(`
{{- range . -}}
%s
{{ else }}nothing found{{ end }}`, dnsRecord))
)

func (tg *Telegram) DNSRecordsCreate(ctx context.Context, res actions.DNSRecordsCreateResult) {
	tg.tplResult(ctx, dnsRecordTemplate, res)
}

func (tg *Telegram) DNSRecordsList(ctx context.Context, res actions.DNSRecordsListResult) {
	tg.tplResult(ctx, dnsRecordsTemplate, res)
}

func (tg *Telegram) DNSRecordsDelete(ctx context.Context, res actions.DNSRecordsDeleteResult) {
	tg.txtResult(ctx, "dns record deleted")
}

//
// HTTP routes
//

var (
	httpRoute = `
{{- $r := . -}}
<b>[{{ $r.Index }}] - </b>{{ $r.Method }} {{ $r.Path }} -> {{ $r.Code }}`

	httpRouteTemplate = tpl(httpRoute)

	httpRoutesTemplate = tpl(fmt.Sprintf(`
{{- range . -}}
%s
{{ else }}nothing found{{ end }}`, httpRoute))
)

func (tg *Telegram) HTTPRoutesCreate(ctx context.Context, res actions.HTTPRoutesCreateResult) {
	tg.tplResult(ctx, httpRouteTemplate, res)
}

func (tg *Telegram) HTTPRoutesList(ctx context.Context, res actions.HTTPRoutesListResult) {
	tg.tplResult(ctx, httpRoutesTemplate, res)
}

func (tg *Telegram) HTTPRoutesDelete(ctx context.Context, res actions.HTTPRoutesDeleteResult) {
	tg.txtResult(ctx, "http route deleted")
}

//
// Users
//

func (tg *Telegram) UsersCreate(ctx context.Context, res actions.UsersCreateResult) {
	tg.txtResult(ctx, fmt.Sprintf("user %q created", res.Name))
}

func (tg *Telegram) UsersDelete(ctx context.Context, res actions.UsersDeleteResult) {
	tg.txtResult(ctx, fmt.Sprintf("user %q deleted", res.Name))
}

//
// Events
//

var (
	event = `
{{- $e := . -}}
<b>[{{ $e.Index }}]</b> - {{ $e.Protocol | upper }} from {{ $e.RemoteAddr }} at {{ $e.ReceivedAt }}`

	eventTemplate = tpl(event + `

{{ $e.RW | b64dec }}
`)

	eventsTemplate = tpl(fmt.Sprintf(`
{{- range . -}}
%s
{{ else }}nothing found{{ end -}}`, event))
)

func (tg *Telegram) EventsList(ctx context.Context, res actions.EventsListResult) {
	tg.tplResult(ctx, eventsTemplate, res)
}

func (tg *Telegram) EventsGet(ctx context.Context, res actions.EventsGetResult) {
}
