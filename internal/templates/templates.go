package templates

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	htmltemplate "html/template"
	texttemplate "text/template"

	"github.com/Masterminds/sprig/v3"

	"github.com/russtone/sonar/internal/actions"
)

type Templates struct {
	templates map[string]Template
	options   options
}

func New(domain string, opts ...Option) *Templates {

	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	return &Templates{
		options: options,
		templates: map[string]Template{
			actions.ProfileGetResultID:       MakeTemplate(profileGet, domain, options.html, options.markup),
			actions.PayloadsListResultID:     MakeTemplate(payloadsList, domain, options.html, options.markup),
			actions.PayloadsCreateResultID:   MakeTemplate(payloadsCreate, domain, options.html, options.markup),
			actions.PayloadsUpdateResultID:   MakeTemplate(payloadsUpdate, domain, options.html, options.markup),
			actions.PayloadsDeleteResultID:   MakeTemplate(payloadsDelete, domain, options.html, options.markup),
			actions.DNSRecordsListResultID:   MakeTemplate(dnsRecordsList, domain, options.html, options.markup),
			actions.DNSRecordsCreateResultID: MakeTemplate(dnsRecordsCreate, domain, options.html, options.markup),
			actions.DNSRecordsDeleteResultID: MakeTemplate(dnsRecordsDelete, domain, options.html, options.markup),
			actions.HTTPRoutesListResultID:   MakeTemplate(httpRoutesList, domain, options.html, options.markup),
			actions.HTTPRoutesCreateResultID: MakeTemplate(httpRoutesCreate, domain, options.html, options.markup),
			actions.HTTPRoutesDeleteResultID: MakeTemplate(httpRoutesDelete, domain, options.html, options.markup),
			actions.EventsListResultID:       MakeTemplate(eventsList, domain, options.html, options.markup),
			actions.EventsGetResultID:        MakeTemplate(eventsGet, domain, options.html, options.markup),
		},
	}
}

func (t *Templates) Execute(res actions.Result) (string, error) {
	tpl, ok := t.templates[res.ResultID()]
	if !ok {
		return "", fmt.Errorf("no template for %q", res.ResultID())
	}

	buf := &bytes.Buffer{}

	if err := tpl.Execute(buf, res); err != nil {
		return "", fmt.Errorf("template error for %q: %v", res.ResultID(), err)
	}

	s := buf.String()

	if t.options.newLine {
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
	}

	return s, nil
}

type Template interface {
	Execute(io.Writer, any) error
}

func MakeTemplate(s string, domain string, html bool, markup map[string]string) Template {

	// Replace all pseudotags like "<bold>" or "<code>" using
	// provided markup map.
	// TODO: maybe move to Execute function, after executing template
	for tag, replacement := range markup {
		s = strings.ReplaceAll(s, tag, replacement)
	}

	extraFuncs := htmltemplate.FuncMap{
		"domain": func() string {
			return domain
		},
	}

	if html {
		return htmltemplate.Must(htmltemplate.
			New("").
			Funcs(sprig.FuncMap()).
			Funcs(extraFuncs).
			Parse(s),
		)
	} else {
		return texttemplate.Must(texttemplate.
			New("").
			Funcs(sprig.FuncMap()).
			Funcs(extraFuncs).
			Parse(s),
		)
	}
}
