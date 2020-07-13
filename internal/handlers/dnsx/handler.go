package dnsx

import (
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/pointer"
)

func (h *DNSX) handleFunc(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) == 0 {
		return
	}

	msg := &dns.Msg{}
	msg.SetReply(r)
	msg.Authoritative = true

	name := r.Question[0].Name
	qtype := r.Question[0].Qtype

	if h.subdomainRegexp.MatchString(name) {
		if payload, rec := h.findDynamicRecords(name, qtype); rec != nil {
			origin := fmt.Sprintf("%s.%s", payload.Subdomain, h.origin)

			switch rec.Strategy {

			case models.DNSStrategyAll:
				msg.Answer = rec.RRs(origin)

			case models.DNSStrategyRoundRobin:
				if rec.LastAnswer != nil {
					msg.Answer = rotate(rec.LastAnswerRRs(h.origin))
				} else {
					msg.Answer = rec.RRs(origin)
				}

			case models.DNSStrategyRebind:
				if rec.LastAnswer != nil &&
					rec.LastAccessedAt != nil &&
					time.Now().UTC().Sub(*rec.LastAccessedAt) < time.Second*3 {
					i := findIndex(rec.Values, rec.LastAnswer[0]) + 1
					rrs := rec.RRs(origin)
					if i < len(rrs) {
						msg.Answer = rrs[i : i+1]
					} else {
						msg.Answer = rrs[len(rrs)-1:]
					}
				} else {
					msg.Answer = rec.RRs(origin)[:1]
				}
			}

			rec.LastAnswer = rrsToStrings(msg.Answer)
			rec.LastAccessedAt = pointer.Time(time.Now().UTC())

			h.db.DNSRecordsUpdate(rec)
		}
	}

	if len(msg.Answer) == 0 {
		msg.Answer = h.findStaticRecords(name, qtype)
	}

	if len(msg.Answer) > 0 {
		msg.Answer = fixWilidcards(msg.Answer, name)
	} else {
		msg.Rcode = dns.RcodeServerFailure
	}

	w.WriteMsg(msg)
}

func (h *DNSX) findDynamicRecords(name string, qtype uint16) (*models.Payload, *models.DNSRecord) {
	// test1.test2.00b18489.sonar.local -> [test1, test2, 00b18489]
	parts := strings.Split(strings.TrimSuffix(name, "."+h.origin+"."), ".")

	if len(parts) < 2 {
		return nil, nil
	}

	// 00b18489
	right := parts[len(parts)-1]

	p, err := h.db.PayloadsGetBySubdomain(right)
	if err != nil {
		return nil, nil
	}

	// test1.test2
	left := strings.Join(parts[0:len(parts)-1], ".")
	names := []string{left}

	// [test1.test2, *.test2, *]
	names = append(names, makeWildcards(left)...)

	for _, n := range names {
		r, err := h.db.DNSRecordsGetByPayloadNameType(p.ID, n, qtypeStr(qtype))
		if err != nil {
			continue
		}

		return p, r
	}

	return nil, nil
}

func (h *DNSX) findStaticRecords(name string, qtype uint16) []dns.RR {
	names := []string{name}
	names = append(names, makeWildcards(name)...)

	for _, n := range names {
		rrs := h.getStatic(n, qtype)
		if rrs == nil {
			continue
		}

		return rrs
	}

	return nil
}
