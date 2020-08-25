package main

import (
	"text/template"

	"github.com/Masterminds/sprig"
)

var (
	payloadTpl = tpl(`<bold>[{{ .Name }}]</> - {{ .Subdomain }}.{{ $.Domain }} ({{ .NotifyProtocols | join ", " }})`)
)

func tpl(tpl string) *template.Template {
	return template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(tpl))
}
