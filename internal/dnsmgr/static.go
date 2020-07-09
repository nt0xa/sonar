package dnsmgr

import "github.com/miekg/dns"

func (mgr *DNSMgr) addStatic(rr dns.RR) {
	key := makeKey(rr.Header().Name, rr.Header().Rrtype)

	if _, ok := mgr.static[key]; !ok {
		mgr.static[key] = make([]dns.RR, 0)
	}

	mgr.static[key] = append(mgr.static[key], rr)
}

func (mgr *DNSMgr) delStatic(name string, qtype uint16) {
	key := makeKey(name, qtype)

	if _, ok := mgr.static[key]; !ok {
		return
	}

	delete(mgr.static, key)
}

func (mgr *DNSMgr) getStatic(name string, qtype uint16) []dns.RR {
	key := makeKey(name, qtype)

	if _, ok := mgr.static[key]; !ok {
		return nil
	}

	return mgr.static[key]
}
