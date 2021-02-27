package dnsx

import (
	"fmt"
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

func fixWildcards(rrs []dns.RR, name string) []dns.RR {
	newRRs := make([]dns.RR, 0)

	for _, rr := range rrs {
		// In case of wildcard it is required to change name
		// as in was in the question.
		if strings.HasPrefix(rr.Header().Name, "*") {
			rr = dns.Copy(rr)
			rr.Header().Name = name
		}

		newRRs = append(newRRs, rr)
	}

	return newRRs
}

func rrsToStrings(rrs []dns.RR) []string {
	res := make([]string, 0)
	for _, rr := range rrs {
		res = append(res, models.DNSRRToString(rr))
	}
	return res
}
