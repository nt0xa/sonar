package service

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	v "github.com/nt0xa/sonar/pkg/valid"
)

var (
	dnsNamePattern  = `[a-z0-9]{1}([a-z0-9-]*[a-z0-9]{1})?`
	subdomainRegexp = regexp.MustCompile(fmt.Sprintf(`^(\*|%[1]s)(\.%[1]s)*$`, dnsNamePattern))
	fqdnRegexp      = regexp.MustCompile(fmt.Sprintf(`^(%s\.)+$`, dnsNamePattern))
)

// subdomain reports whether s is a valid subdomain (wildcards allowed).
func subdomain(s string) error {
	if !subdomainRegexp.MatchString(s) {
		return errors.New("invalid subdomain")
	}
	return nil
}

// fqdn reports whether s is a fully qualified domain name (trailing dot).
func fqdn(s string) error {
	if !fqdnRegexp.MatchString(s) {
		return errors.New("invalid fqdn")
	}
	return nil
}

// mx reports whether s is a valid MX record ("<priority> <fqdn>").
func mx(s string) error {
	parts := strings.Split(s, " ")
	if len(parts) == 2 {
		if _, err := strconv.Atoi(parts[0]); err == nil && fqdnRegexp.MatchString(parts[1]) {
			return nil
		}
	}
	return errors.New("invalid mx record")
}

// caa reports whether s is a valid CAA record ("<flag> <tag> <value>").
func caa(s string) error {
	var (
		flag uint8
		tag  string
		val  string
	)
	if _, err := fmt.Sscanf(s, "%d %s %q", &flag, &tag, &val); err != nil {
		return fmt.Errorf("invalid caa record: %w", err)
	}
	return nil
}

// ip4 reports whether s is a valid IPv4 address.
func ip4(s string) error {
	if ip := net.ParseIP(s); ip == nil || ip.To4() == nil {
		return errors.New("must be a valid IPv4 address")
	}
	return nil
}

// ip6 reports whether s is a valid IPv6 address.
func ip6(s string) error {
	if ip := net.ParseIP(s); ip == nil || ip.To4() != nil {
		return errors.New("must be a valid IPv6 address")
	}
	return nil
}

// dnsValueRule returns the per-value validation rule for a DNS record type.
func dnsValueRule(t DNSRecordType) v.StringRule {
	switch t {
	case DNSRecordTypeA:
		return v.By(ip4)
	case DNSRecordTypeAAAA:
		return v.By(ip6)
	case DNSRecordTypeMX:
		return v.By(mx)
	case DNSRecordTypeCNAME:
		return v.By(fqdn)
	case DNSRecordTypeCAA:
		return v.By(caa)
	default:
		return v.Required
	}
}
