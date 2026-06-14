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

	"github.com/nt0xa/sonar/internal/modules"
	"github.com/nt0xa/sonar/internal/service"
)

type Templates struct {
	options options

	// Result templates, keyed by the service output Go type in RenderResult.
	profileGet       *template
	payload          *template // payloads create/update
	payloadsList     *template
	payloadsDelete   *template
	payloadsClear    *template
	dnsRecord        *template // dns records create
	dnsRecordsList   *template
	dnsRecordsDelete *template
	dnsRecordsClear  *template
	httpRoute        *template // http routes create/update
	httpRoutesList   *template
	httpRoutesDelete *template
	httpRoutesClear  *template
	usersCreate      *template
	usersDelete      *template
	eventsGet        *template
	eventsList       *template
	auditRecordsList *template
	auditRecordsGet  *template

	// Notification templates, still id-keyed via the PerTemplate option API.
	notificationHeader *template
	notificationBody   *template
}

func New(domain string, opts ...Option) *Templates {
	options := options{
		defaultOptions: defaultTemplateOptions(),
		perTemplate:    make(map[string][]TemplateOption),
	}

	for _, opt := range opts {
		opt(&options)
	}

	// Result templates all use the default options; only notifications have
	// per-template overrides.
	mk := func(content string) *template {
		return makeTemplate(content, domain, options.defaultOptions)
	}

	return &Templates{
		options:            options,
		profileGet:         mk(profileGet),
		payload:            mk(payload),
		payloadsList:       mk(payloadsList),
		payloadsDelete:     mk(payloadsDelete),
		payloadsClear:      mk(payloadsClear),
		dnsRecord:          mk(dnsRecord),
		dnsRecordsList:     mk(dnsRecordsList),
		dnsRecordsDelete:   mk(dnsRecordsDelete),
		dnsRecordsClear:    mk(dnsRecordsClear),
		httpRoute:          mk(httpRoute),
		httpRoutesList:     mk(httpRoutesList),
		httpRoutesDelete:   mk(httpRoutesDelete),
		httpRoutesClear:    mk(httpRoutesClear),
		usersCreate:        mk(usersCreate),
		usersDelete:        mk(usersDelete),
		eventsGet:          mk(eventsGet),
		eventsList:         mk(eventsList),
		auditRecordsList:   mk(auditRecordsList),
		auditRecordsGet:    mk(auditRecordsGet),
		notificationHeader: makeTemplate(notificationHeader, domain, options.get(NotificationHeaderID)),
		notificationBody:   makeTemplate(notificationBody, domain, options.get(NotificationBodyID)),
	}
}

// RenderResult renders a service command output, picking the template by the
// output's concrete Go type.
func (t *Templates) RenderResult(out any) (string, error) {
	var tpl *template

	switch out.(type) {
	case *service.ProfileGetOutput:
		tpl = t.profileGet
	case *service.PayloadsCreateOutput, *service.PayloadsUpdateOutput:
		tpl = t.payload
	case service.PayloadsListOutput:
		tpl = t.payloadsList
	case *service.PayloadsDeleteOutput:
		tpl = t.payloadsDelete
	case service.PayloadsClearOutput:
		tpl = t.payloadsClear
	case *service.DNSRecordsCreateOutput:
		tpl = t.dnsRecord
	case service.DNSRecordsListOutput:
		tpl = t.dnsRecordsList
	case *service.DNSRecordsDeleteOutput:
		tpl = t.dnsRecordsDelete
	case service.DNSRecordsClearOutput:
		tpl = t.dnsRecordsClear
	case *service.HTTPRoutesCreateOutput, *service.HTTPRoutesUpdateOutput:
		tpl = t.httpRoute
	case service.HTTPRoutesListOutput:
		tpl = t.httpRoutesList
	case *service.HTTPRoutesDeleteOutput:
		tpl = t.httpRoutesDelete
	case service.HTTPRoutesClearOutput:
		tpl = t.httpRoutesClear
	case *service.UsersCreateOutput:
		tpl = t.usersCreate
	case *service.UsersDeleteOutput:
		tpl = t.usersDelete
	case *service.EventsGetOutput:
		tpl = t.eventsGet
	case service.EventsListOutput:
		tpl = t.eventsList
	case service.AuditRecordsListOutput:
		tpl = t.auditRecordsList
	case *service.AuditRecordsGetOutput:
		tpl = t.auditRecordsGet
	default:
		return "", fmt.Errorf("no template for %T", out)
	}

	return tpl.render(out)
}

func (t *Templates) RenderNotification(n *modules.Notification) (string, string, error) {
	header, err := t.notificationHeader.render(n)
	if err != nil {
		return "", "", err
	}

	body, err := t.notificationBody.render(n)
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
		"flag": FlagEmoji,
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

func FlagEmoji(countryCode string) string {
	countryCode = strings.ToUpper(countryCode)
	if len(countryCode) != 2 {
		return "" // Invalid code
	}
	runes := []rune{}
	for _, c := range countryCode {
		if c < 'A' || c > 'Z' {
			return "" // Invalid character
		}
		runes = append(runes, 0x1F1E6+(c-'A'))
	}
	return string(runes)
}
