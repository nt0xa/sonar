package database

import "github.com/miekg/dns"

// Qtype maps the DNSRecord's Type to a dns.Type uint16 value.
func (r DNSRecord) Qtype() uint16 {
	switch r.Type {
	case DNSRecordTypeA:
		return dns.TypeA
	case DNSRecordTypeAAAA:
		return dns.TypeAAAA
	case DNSRecordTypeMX:
		return dns.TypeMX
	case DNSRecordTypeTXT:
		return dns.TypeTXT
	case DNSRecordTypeCNAME:
		return dns.TypeCNAME
	case DNSRecordTypeNS:
		return dns.TypeNS
	case DNSRecordTypeCAA:
		return dns.TypeCAA
	default:
		return dns.TypeNone
	}
}

// TODO: cleanup
var DNSTypesAll = func() []string {
	vals := AllDNSRecordTypeValues()
	res := make([]string, len(vals))
	for i, v := range vals {
		res[i] = string(v)
	}
	return res
}()

// TODO: cleanup
var DNSStrategiesAll = func() []string {
	vals := AllDNSStrategyValues()
	res := make([]string, len(vals))
	for i, v := range vals {
		res[i] = string(v)
	}
	return res
}()
