package telegram

import (
	"html/template"

	"github.com/Masterminds/sprig"
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

	listPayloadTemplate = tpl(`{{range .Payloads}}<b>[{{ .Name }}]</b> - <code>{{ .Subdomain }}.{{ $.Domain }}</code> ({{ .NotifyProtocols | join ", " }})
{{else}}you don't have any payloads yet{{end}}`)

	dnsRecordTemplate = tpl(`{{range .RRs}}<code>{{ . }}</code>
{{end}}`)

	userTemplate = tpl("" +
		"<b>Telegram ID:</b> <code>{{ .TelegramID }}</code>\n" +
		"<b>API token:</b> <code>{{ .APIToken }}</code>",
	)
)

func tpl(tpl string) *template.Template {
	return template.Must(template.New("msg").Funcs(sprig.FuncMap()).Parse(tpl))
}
