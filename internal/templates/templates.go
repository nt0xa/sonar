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
	"github.com/russtone/sonar/internal/modules"
)

type Templates struct {
	options options

	results map[string]Template

	notificationHeader Template
	notificationBody   Template
}

func New(domain string, opts ...Option) *Templates {

	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	return &Templates{
		options: options,
		results: map[string]Template{
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
		notificationHeader: MakeTemplate(notificationHeader, domain, options.html, options.markup),
		notificationBody:   MakeTemplate(notificationBody, domain, options.html, options.markup),
	}
}

func (t *Templates) RenderResult(res actions.Result) (string, error) {
	tpl, ok := t.results[res.ResultID()]
	if !ok {
		return "", fmt.Errorf("no template for %q", res.ResultID())
	}

	return t.render(tpl, res)
}

func (t *Templates) RenderNotification(n *modules.Notification) (string, string, error) {
	header, err := t.render(t.notificationHeader, n)
	if err != nil {
		return "", "", err
	}

	body, err := t.render(t.notificationBody, n)
	if err != nil {
		return "", "", err
	}

	return header, body, nil
}

func (t *Templates) render(tpl Template, data any) (string, error) {
	buf := &bytes.Buffer{}

	if err := tpl.Execute(buf, data); err != nil {
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
