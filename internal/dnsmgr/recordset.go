package dnsmgr

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

type recordSet struct {
	records map[string]map[uint16][]dns.RR
}

func newRecordSet() *recordSet {
	return &recordSet{
		records: make(map[string]map[uint16][]dns.RR),
	}
}

func (rs *recordSet) getByName(name string) map[uint16][]dns.RR {
	name = strings.ToLower(name)

	if _, ok := rs.records[name]; !ok {
		return nil
	}

	return rs.records[name]
}

func (rs *recordSet) getByNameAndQtype(name string, qtype uint16) []dns.RR {
	name = strings.ToLower(name)

	if _, ok := rs.records[name]; !ok {
		return nil
	}

	if _, ok := rs.records[name][qtype]; !ok {
		return nil
	}

	return rs.records[name][qtype]
}

func (rs *recordSet) add(rr dns.RR) {
	name := strings.ToLower(rr.Header().Name)

	var qtmap map[uint16][]dns.RR

	if m, ok := rs.records[name]; !ok {
		qtmap = make(map[uint16][]dns.RR)
		rs.records[name] = qtmap
	} else {
		qtmap = m
	}

	var rrs []dns.RR

	qtype := rr.Header().Rrtype

	if values, ok := qtmap[qtype]; ok {
		values = append(values, rr)
		rrs = values
	} else {
		rrs = make([]dns.RR, 0)
		rrs = append(rrs, rr)
	}

	rs.records[name][qtype] = rrs
}

func (rs *recordSet) delByName(name string) {
	name = strings.ToLower(name)

	if _, ok := rs.records[name]; !ok {
		return
	}

	delete(rs.records, name)
}

func (rs *recordSet) delByNameAndQtype(name string, qtype uint16) {
	name = strings.ToLower(name)

	if _, ok := rs.records[name]; !ok {
		return
	}

	if _, ok := rs.records[name][qtype]; !ok {
		return
	}

	delete(rs.records[name], qtype)
}

func (rs *recordSet) String() string {
	s := ""

	for name, qtmap := range rs.records {
		s += fmt.Sprintf("%s:\n", name)

		for qtype, rrs := range qtmap {
			s += fmt.Sprintf("\t%s:\n", qtypeStr(qtype))

			for _, rr := range rrs {
				switch r := rr.(type) {
				case *dns.A:
					s += fmt.Sprintf("\t\t%s\n", r.A)
				case *dns.AAAA:
					s += fmt.Sprintf("\t\t%s\n", r.AAAA)
				case *dns.MX:
					s += fmt.Sprintf("\t\t%d %s\n", r.Preference, r.Mx)
				case *dns.TXT:
					s += fmt.Sprintf("\t\t%s\n", r.Txt)
				}
			}
		}
	}

	return s
}
