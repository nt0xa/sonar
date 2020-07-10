package dnsmgr

import (
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/pointer"
)

func (mgr *DNSMgr) HandleFunc(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) == 0 {
		return
	}

	msg := &dns.Msg{}
	msg.SetReply(r)

	name := r.Question[0].Name
	qtype := r.Question[0].Qtype

	if mgr.dynamicRegex.MatchString(name) {
		if rec := mgr.findDynamicRecords(name, qtype); rec != nil {
			switch rec.Strategy {

			case models.DNSStrategyAll:
				msg.Answer = rec.RRs(mgr.origin)

			case models.DNSStrategyRoundRobin:
				if rec.LastAnswer != nil {
					msg.Answer = rotate(rec.LastAnswerRRs(mgr.origin))
				} else {
					msg.Answer = rec.RRs(mgr.origin)
				}

			case models.DNSStrategyRebind:
				if rec.LastAnswer != nil &&
					rec.LastAccessedAt != nil &&
					time.Now().UTC().Sub(*rec.LastAccessedAt) < time.Second*3 {
					i := findIndex(rec.Values, rec.LastAnswer[0]) + 1
					rrs := rec.RRs(mgr.origin)
					if i < len(rrs) {
						msg.Answer = rrs[i : i+1]
					} else {
						msg.Answer = rrs[len(rrs)-1:]
					}
				} else {
					msg.Answer = rec.RRs(mgr.origin)[:1]
				}
			}

			rec.LastAnswer = rrsToStrings(msg.Answer)
			rec.LastAccessedAt = pointer.Time(time.Now().UTC())

			mgr.db.DNSRecordsUpdate(rec)
		}
	}

	if len(msg.Answer) == 0 {
		msg.Answer = mgr.findStaticRecords(name, qtype)
	}

	if len(msg.Answer) > 0 {
		msg.Answer = fixWilidcards(msg.Answer, name)
	} else {
		msg.Rcode = dns.RcodeServerFailure
	}

	w.WriteMsg(msg)
}

func (mgr *DNSMgr) findDynamicRecords(name string, qtype uint16) *models.DNSRecord {
	// test1.test2.00b18489.sonar.local -> [test1 test2 00b18489]
	parts := strings.Split(strings.TrimSuffix(name, "."+mgr.origin+"."), ".")

	if len(parts) < 2 {
		return nil
	}

	// 00b18489
	right := parts[len(parts)-1]

	p, err := mgr.db.PayloadsGetBySubdomain(right)
	if err != nil {
		return nil
	}

	// test1.test2
	left := strings.Join(parts[0:len(parts)-1], ".")
	names := []string{left}

	// [test1.test2 *.test2 *]
	names = append(names, makeWildcards(left)...)

	for _, n := range names {
		r, err := mgr.db.DNSRecordsGetByPayloadNameType(p.ID, n, qtypeStr(qtype))
		if err != nil {
			continue
		}

		return r
	}

	return nil
}

func (mgr *DNSMgr) findStaticRecords(name string, qtype uint16) []dns.RR {
	names := []string{name}
	names = append(names, makeWildcards(name)...)

	for _, n := range names {
		rrs := mgr.getStatic(n, qtype)
		if rrs == nil {
			continue
		}

		return rrs
	}

	return nil
}
