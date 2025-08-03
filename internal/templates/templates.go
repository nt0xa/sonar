package templates

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"strings"

	htmltemplate "html/template"
	texttemplate "text/template"

	"github.com/Masterminds/sprig/v3"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/modules"
)

type Templates struct {
	options   options
	templates map[string]*template
}

func New(domain string, opts ...Option) *Templates {
	options := options{
		defaultOptions: defaultTemplateOptions(),
		perTemplate:    make(map[string]templateOptions),
	}

	for _, opt := range opts {
		opt(&options)
	}

	templates := make(map[string]*template)

	for id, content := range templatesMap {
		templates[id] = makeTemplate(content, domain, options.get(id))
	}

	return &Templates{
		options:   options,
		templates: templates,
	}
}

func (t *Templates) RenderResult(res actions.Result) (string, error) {
	tpl, ok := t.templates[res.ResultID()]
	if !ok {
		return "", fmt.Errorf("no template for %q", res.ResultID())
	}

	return tpl.render(res)
}

func (t *Templates) RenderNotification(n *modules.Notification) (string, string, error) {
	header, err := t.templates[NotificationHeaderID].render(n)
	if err != nil {
		return "", "", err
	}

	body, err := t.templates[NotificationBodyID].render(n)
	if err != nil {
		return "", "", err
	}

	return header, body, nil
}

type Template interface {
	Execute(io.Writer, any) error
}

type template struct {
	options templateOptions
	tpl     Template
}

func makeTemplate(s string, domain string, opts templateOptions) *template {

	// Replace all pseudotags like "<bold>" or "<code>" using
	// provided markup map.
	// TODO: maybe move to Execute function, after executing template
	for tag, replacement := range opts.markup {
		s = strings.ReplaceAll(s, tag, replacement)
	}

	extraFuncs := htmltemplate.FuncMap{
		"domain": func() string {
			return domain
		},
	}

	maps.Copy(extraFuncs, opts.extraFuncs)

	t := &template{options: opts}

	if opts.html {
		t.tpl = htmltemplate.Must(htmltemplate.
			New("").
			Funcs(sprig.FuncMap()).
			Funcs(extraFuncs).
			Parse(s),
		)
	} else {
		t.tpl = texttemplate.Must(texttemplate.
			New("").
			Funcs(sprig.FuncMap()).
			Funcs(extraFuncs).
			Parse(s),
		)
	}

	return t
}

func (t *template) render(data any) (string, error) {
	buf := &bytes.Buffer{}

	if err := t.tpl.Execute(buf, data); err != nil {
		return "", fmt.Errorf("template error: %v", err)
	}

	s := buf.String()

	if t.options.newLine {
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
	}

	return s, nil
}
