package dnsmgr

import (
	"fmt"
	"io"
	"strings"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/miekg/dns"
)

func qtypeStr(qtype uint16) string {
	return dns.Type(qtype).String()
}

func makeKey(name string, qtype uint16) string {
	return fmt.Sprintf("%s/%s", strings.ToLower(name), qtypeStr(qtype))
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

func fixWilidcards(rrs []dns.RR, name string) []dns.RR {
	newRRs := make([]dns.RR, 0)

	for _, rr := range rrs {
		// In case of wildcard is is required to change name
		// as in was in the question.
		if strings.HasPrefix(rr.Header().Name, "*") {
			rr = dns.Copy(rr)
			rr.Header().Name = name
		}

		newRRs = append(newRRs, rr)
	}

	return newRRs
}

func rotate(rrs []dns.RR) []dns.RR {
	if len(rrs) <= 1 {
		return rrs
	}

	newRRs := make([]dns.RR, len(rrs))

	copy(newRRs, rrs[1:])

	newRRs[len(rrs)-1] = rrs[0]

	return newRRs
}

func rrsToStrings(rrs []dns.RR) []string {
	res := make([]string, 0)
	for _, rr := range rrs {
		res = append(res, models.DNSRRToString(rr))
	}

	return res
}

func findIndex(values []string, value string) int {
	for i, v := range values {
		if value == v {
			return i
		}
	}

	return -1
}
