package dnsmgr

import (
	"strings"

	"github.com/miekg/dns"
)

func (mgr *DNSMgr) HandleFunc(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) == 0 {
		return
	}

	name := r.Question[0].Name
	qtype := r.Question[0].Qtype

	var rrs []dns.RR

	if mgr.dynamicRecordRegex.MatchString(name) {
		rrs = mgr.findDynamicRecords(name, qtype)
	}

	if len(rrs) == 0 {
		rrs = mgr.findStaticRecords(name, qtype)
	}

	msg := &dns.Msg{}
	msg.SetReply(r)

	if len(rrs) > 0 {
		for _, rr := range rrs {
			// In case of wildcard is is required to change name
			// as in was in the question.
			if strings.HasPrefix(rr.Header().Name, "*") {
				rr = dns.Copy(rr)
				rr.Header().Name = name
			}

			msg.Answer = append(msg.Answer, rr)
		}
	} else {
		msg.Rcode = dns.RcodeServerFailure
	}

	w.WriteMsg(msg)
}

func (mgr *DNSMgr) findDynamicRecords(name string, qtype uint16) []dns.RR {
	parts := strings.Split(strings.TrimSuffix(name, "."+mgr.origin+"."), ".")

	if len(parts) < 2 {
		return nil
	}

	right := parts[len(parts)-1]

	p, err := mgr.db.PayloadsGetBySubdomain(right)
	if err != nil {
		return nil
	}

	left := strings.Join(parts[0:len(parts)-1], ".")
	names := []string{left}
	names = append(names, makeWildcards(left)...)

	for _, n := range names {
		r, err := mgr.db.DNSRecordsGetByPayloadNameType(p.ID, n, qtypeStr(qtype))
		if err != nil {
			continue
		}
		return r.RRs(mgr.origin)
	}

	return nil
}

func (mgr *DNSMgr) findStaticRecords(name string, qtype uint16) []dns.RR {
	names := []string{name}
	names = append(names, makeWildcards(name)...)

	for _, n := range names {
		rrs := mgr.staticRecords.getByNameAndQtype(n, qtype)
		if rrs == nil {
			continue
		}

		return rrs
	}

	return nil
}
