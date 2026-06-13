package cmd2

import (
	"strings"

	"github.com/spf13/pflag"

	"github.com/nt0xa/sonar/internal/service"
)

// These bridges let pflag's VarP bind directly into the typed enum fields of the
// service inputs. Each delegates to the generated service.Parse<Enum>.
var (
	_ pflag.Value = dnsTypeValue{}
	_ pflag.Value = dnsStrategyValue{}
	_ pflag.Value = httpMethodValue{}
	_ pflag.Value = auditResourceTypeValue{}
	_ pflag.Value = auditActionValue{}
	_ pflag.Value = (*protoSlice)(nil)
)

type dnsTypeValue struct{ p *service.DNSRecordType }

func (v dnsTypeValue) String() string { return string(*v.p) }
func (v dnsTypeValue) Type() string   { return "type" }
func (v dnsTypeValue) Set(s string) error {
	t, err := service.ParseDNSRecordType(s)
	if err != nil {
		return err
	}
	*v.p = t
	return nil
}

type dnsStrategyValue struct{ p *service.DNSRecordStrategy }

func (v dnsStrategyValue) String() string { return string(*v.p) }
func (v dnsStrategyValue) Type() string   { return "strategy" }
func (v dnsStrategyValue) Set(s string) error {
	st, err := service.ParseDNSRecordStrategy(s)
	if err != nil {
		return err
	}
	*v.p = st
	return nil
}

type httpMethodValue struct{ p *service.HTTPMethod }

func (v httpMethodValue) String() string { return string(*v.p) }
func (v httpMethodValue) Type() string   { return "method" }
func (v httpMethodValue) Set(s string) error {
	m, err := service.ParseHTTPMethod(s)
	if err != nil {
		return err
	}
	*v.p = m
	return nil
}

type auditResourceTypeValue struct{ p *service.AuditResourceType }

func (v auditResourceTypeValue) String() string { return string(*v.p) }
func (v auditResourceTypeValue) Type() string   { return "resource-type" }
func (v auditResourceTypeValue) Set(s string) error {
	rt, err := service.ParseAuditResourceType(s)
	if err != nil {
		return err
	}
	*v.p = rt
	return nil
}

type auditActionValue struct{ p *service.AuditAction }

func (v auditActionValue) String() string { return string(*v.p) }
func (v auditActionValue) Type() string   { return "action" }
func (v auditActionValue) Set(s string) error {
	a, err := service.ParseAuditAction(s)
	if err != nil {
		return err
	}
	*v.p = a
	return nil
}

// protoSlice binds a []ProtoCategory flag. Mirrors pflag's stringSlice: the first
// Set replaces, later occurrences append; a single value may be comma-separated.
type protoSlice struct {
	p       *[]service.ProtoCategory
	changed bool
}

func (v *protoSlice) Type() string { return "strings" }

func (v *protoSlice) String() string {
	if v.p == nil || len(*v.p) == 0 {
		return ""
	}
	parts := make([]string, len(*v.p))
	for i, c := range *v.p {
		parts[i] = string(c)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func (v *protoSlice) Set(s string) error {
	parts := strings.Split(s, ",")
	out := make([]service.ProtoCategory, 0, len(parts))
	for _, p := range parts {
		c, err := service.ParseProtoCategory(strings.TrimSpace(p))
		if err != nil {
			return err
		}
		out = append(out, c)
	}
	if !v.changed {
		*v.p = out
		v.changed = true
	} else {
		*v.p = append(*v.p, out...)
	}
	return nil
}
