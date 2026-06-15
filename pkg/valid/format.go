package valid

import (
	"errors"
	"net"
	"net/url"
	"regexp"
)

// domainRegexp matches a DNS hostname like "example.com": dot-separated labels
// of letters/digits/hyphens (not starting or ending with a hyphen) and a
// letters-only TLD.
var domainRegexp = regexp.MustCompile(
	`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`,
)

// IP asserts the string is a valid IPv4 or IPv6 address.
func IP(s string) error {
	if net.ParseIP(s) == nil {
		return errors.New("must be a valid IP address")
	}
	return nil
}

// Domain asserts the string is a valid domain name (e.g. example.com).
func Domain(s string) error {
	if len(s) > 255 || !domainRegexp.MatchString(s) {
		return errors.New("must be a valid domain")
	}
	return nil
}

// URL asserts the string is a valid absolute URL with a scheme and host.
func URL(s string) error {
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("must be a valid URL")
	}
	return nil
}
