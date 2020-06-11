package telegram

import "html/template"

func tpl(tpl string) *template.Template {
	return template.Must(template.New("msg").Parse(tpl))
}
