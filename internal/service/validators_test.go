package service

import "testing"

func TestDomainValidators(t *testing.T) {
	tests := []struct {
		name string
		fn   func(string) error
		ok   string
		bad  string
	}{
		{"subdomain", subdomain, "foo.bar", "FOO_BAR"},
		{"fqdn", fqdn, "foo.bar.", "foo.bar"},
		{"mx", mx, "10 mail.example.com.", "mail.example.com"},
		{"ip4", ip4, "1.2.3.4", "::1"},
		{"ip6", ip6, "::1", "1.2.3.4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(tt.ok); err != nil {
				t.Errorf("%s(%q) = %v, want nil", tt.name, tt.ok, err)
			}
			if err := tt.fn(tt.bad); err == nil {
				t.Errorf("%s(%q) = nil, want error", tt.name, tt.bad)
			}
		})
	}
}

func TestDNSRecordsCreateInput_Validate(t *testing.T) {
	valid := DNSRecordsCreateInput{
		PayloadName: "p",
		Name:        "sub",
		Type:        DNSRecordTypeA,
		Values:      []string{"1.2.3.4"},
		Strategy:    DNSRecordStrategyAll,
	}
	if got := valid.Validate(); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}

	t.Run("required and per-type value", func(t *testing.T) {
		in := valid
		in.PayloadName = ""
		in.Values = []string{"not-an-ip"}
		got := in.Validate()
		if got["payloadName"] != "cannot be blank" {
			t.Errorf("payloadName: got %q", got["payloadName"])
		}
		if got["values"] != "element #0: must be a valid IPv4 address" {
			t.Errorf("values: got %q", got["values"])
		}
	})

	t.Run("invalid enum", func(t *testing.T) {
		in := valid
		in.Type = "BOGUS"
		if got := in.Validate(); got["type"] == "" {
			t.Errorf("expected type problem, got %v", got)
		}
	})
}
