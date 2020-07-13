package tpl

import (
	"bytes"
	"html/template"
)

func MustParse(tpl string) *template.Template {
	return template.Must(template.New("tpl").Parse(tpl))
}

func RenderToString(tpl *template.Template, data interface{}) (string, error) {
	var bb bytes.Buffer

	if err := tpl.Execute(&bb, data); err != nil {
		return "", err
	}

	return bb.String(), nil
}
