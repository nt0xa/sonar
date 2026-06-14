package dbsvc

import (
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

func dnsRecord(m database.DNSRecord, payloadSubdomain string) *service.DNSRecord {
	return &service.DNSRecord{
		Index:            int64(m.Index),
		PayloadSubdomain: payloadSubdomain,
		Name:             m.Name,
		Type:             service.DNSRecordType(m.Type),
		TTL:              m.TTL,
		Values:           m.Values,
		Strategy:         service.DNSRecordStrategy(m.Strategy),
		CreatedAt:        m.CreatedAt,
	}
}
