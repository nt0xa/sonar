package mcpserver

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"unicode"
	"unicode/utf8"

	"github.com/nt0xa/sonar/internal/service"
)

// client wraps [service.Service] for the model: it expands payload subdomains
// to full hostnames and converts base64 bodies and data to and from plain text.
type client struct {
	svc    service.Service
	domain string
}

// Payloads

func (c client) PayloadsList(ctx context.Context, in service.PayloadsListInput) ([]service.Payload, error) {
	out, err := c.svc.PayloadsList(ctx, in)
	if err != nil {
		return nil, err
	}
	return mapSlice(out, c.payload), nil
}

func (c client) PayloadsCreate(ctx context.Context, in service.PayloadsCreateInput) (service.Payload, error) {
	out, err := c.svc.PayloadsCreate(ctx, in)
	if err != nil {
		return service.Payload{}, err
	}
	return c.payload(service.Payload(*out)), nil
}

func (c client) PayloadsUpdate(ctx context.Context, in service.PayloadsUpdateInput) (service.Payload, error) {
	out, err := c.svc.PayloadsUpdate(ctx, in)
	if err != nil {
		return service.Payload{}, err
	}
	return c.payload(service.Payload(*out)), nil
}

func (c client) PayloadsDelete(ctx context.Context, in service.PayloadsDeleteInput) (service.Payload, error) {
	out, err := c.svc.PayloadsDelete(ctx, in)
	if err != nil {
		return service.Payload{}, err
	}
	return c.payload(service.Payload(*out)), nil
}

func (c client) PayloadsClear(ctx context.Context, in service.PayloadsClearInput) ([]service.Payload, error) {
	out, err := c.svc.PayloadsClear(ctx, in)
	if err != nil {
		return nil, err
	}
	return mapSlice(out, c.payload), nil
}

// DNS records

func (c client) DNSRecordsList(ctx context.Context, in service.DNSRecordsListInput) ([]service.DNSRecord, error) {
	out, err := c.svc.DNSRecordsList(ctx, in)
	if err != nil {
		return nil, err
	}
	return mapSlice(out, c.dnsRecord), nil
}

func (c client) DNSRecordsCreate(ctx context.Context, in service.DNSRecordsCreateInput) (service.DNSRecord, error) {
	out, err := c.svc.DNSRecordsCreate(ctx, in)
	if err != nil {
		return service.DNSRecord{}, err
	}
	return c.dnsRecord(service.DNSRecord(*out)), nil
}

func (c client) DNSRecordsDelete(ctx context.Context, in service.DNSRecordsDeleteInput) (service.DNSRecord, error) {
	out, err := c.svc.DNSRecordsDelete(ctx, in)
	if err != nil {
		return service.DNSRecord{}, err
	}
	return c.dnsRecord(service.DNSRecord(*out)), nil
}

func (c client) DNSRecordsClear(ctx context.Context, in service.DNSRecordsClearInput) ([]service.DNSRecord, error) {
	out, err := c.svc.DNSRecordsClear(ctx, in)
	if err != nil {
		return nil, err
	}
	return mapSlice(out, c.dnsRecord), nil
}

// HTTP routes (bodies are plain text in both directions)

func (c client) HTTPRoutesList(ctx context.Context, in service.HTTPRoutesListInput) ([]service.HTTPRoute, error) {
	out, err := c.svc.HTTPRoutesList(ctx, in)
	if err != nil {
		return nil, err
	}
	return mapSlice(out, c.httpRoute), nil
}

func (c client) HTTPRoutesCreate(ctx context.Context, in service.HTTPRoutesCreateInput) (service.HTTPRoute, error) {
	in.Body = encode(in.Body)
	out, err := c.svc.HTTPRoutesCreate(ctx, in)
	if err != nil {
		return service.HTTPRoute{}, err
	}
	return c.httpRoute(service.HTTPRoute(*out)), nil
}

func (c client) HTTPRoutesUpdate(ctx context.Context, in service.HTTPRoutesUpdateInput) (service.HTTPRoute, error) {
	if in.Body != nil {
		b := encode(*in.Body)
		in.Body = &b
	}
	out, err := c.svc.HTTPRoutesUpdate(ctx, in)
	if err != nil {
		return service.HTTPRoute{}, err
	}
	return c.httpRoute(service.HTTPRoute(*out)), nil
}

func (c client) HTTPRoutesDelete(ctx context.Context, in service.HTTPRoutesDeleteInput) (service.HTTPRoute, error) {
	out, err := c.svc.HTTPRoutesDelete(ctx, in)
	if err != nil {
		return service.HTTPRoute{}, err
	}
	return c.httpRoute(service.HTTPRoute(*out)), nil
}

func (c client) HTTPRoutesClear(ctx context.Context, in service.HTTPRoutesClearInput) ([]service.HTTPRoute, error) {
	out, err := c.svc.HTTPRoutesClear(ctx, in)
	if err != nil {
		return nil, err
	}
	return mapSlice(out, c.httpRoute), nil
}

// Events

// EventsList omits raw data; use EventsGet for it.
func (c client) EventsList(ctx context.Context, in service.EventsListInput) ([]service.Event, error) {
	out, err := c.svc.EventsList(ctx, in)
	if err != nil {
		return nil, err
	}
	return mapSlice(out, func(e service.Event) service.Event {
		e.Data = nil
		return event(e)
	}), nil
}

func (c client) EventsGet(ctx context.Context, in service.EventsGetInput) (service.Event, error) {
	out, err := c.svc.EventsGet(ctx, in)
	if err != nil {
		return service.Event{}, err
	}
	return event(service.Event(*out)), nil
}

// Conversions

func (c client) payload(p service.Payload) service.Payload {
	p.Subdomain += "." + c.domain
	return p
}

func (c client) dnsRecord(d service.DNSRecord) service.DNSRecord {
	d.PayloadSubdomain += "." + c.domain
	return d
}

func (c client) httpRoute(h service.HTTPRoute) service.HTTPRoute {
	h.PayloadSubdomain += "." + c.domain
	h.Body = decode(h.Body)
	return h
}

func event(e service.Event) service.Event {
	e.Data = mapSlice(e.Data, decode)
	if e.Meta.HTTP != nil {
		h := *e.Meta.HTTP
		h.Request.Body = decode(h.Request.Body)
		h.Response.Body = decode(h.Response.Body)
		e.Meta.HTTP = &h
	}
	return e
}

func encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// decode returns base64 content as text, or as a hexdump if it is binary.
func decode(s string) string {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return s
	}
	if !utf8.Valid(b) {
		return hex.Dump(b)
	}
	for _, r := range string(b) {
		if !unicode.IsPrint(r) && !unicode.IsSpace(r) {
			return hex.Dump(b)
		}
	}
	return string(b)
}

// mapSlice also accepts named slice types such as [service.PayloadsListOutput].
func mapSlice[S ~[]T, T, U any](s S, f func(T) U) []U {
	if s == nil {
		return nil
	}
	out := make([]U, len(s))
	for i, v := range s {
		out[i] = f(v)
	}
	return out
}
