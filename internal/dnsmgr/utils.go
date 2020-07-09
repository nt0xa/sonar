package dnsmgr

import (
	"io"
	"strings"

	"github.com/miekg/dns"
)

func qtypeStr(qtype uint16) string {
	return dns.Type(qtype).String()
}

func parseZoneFile(rdr io.Reader, origin string) ([]dns.RR, error) {
	rrs := make([]dns.RR, 0)
	zp := dns.NewZoneParser(rdr, origin, "")

	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		rrs = append(rrs, rr)
	}

	if err := zp.Err(); err != nil {
		return nil, err
	}

	return rrs, nil
}

func makeWildcard(fqdn string) string {
	split := strings.SplitAfterN(fqdn, ".", 2)
	split[0] = "*"
	return strings.Join(split, ".")
}

func makeWildcards(fqdn string) []string {
	res := make([]string, 0)
	for off, end := 0, false; !end; off, end = dns.NextLabel(fqdn, off) {
		res = append(res, makeWildcard(fqdn[off:]))
	}

	return res
}
