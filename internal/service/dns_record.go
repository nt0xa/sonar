package service

import (
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

func dnsRecord(m database.DNSRecord, payloadSubdomain string) *types.DNSRecord {
	return &types.DNSRecord{
		Index:            int64(m.Index),
		PayloadSubdomain: payloadSubdomain,
		Name:             m.Name,
		Type:             types.DNSRecordType(m.Type),
		TTL:              m.TTL,
		Values:           m.Values,
		Strategy:         types.DNSRecordStrategy(m.Strategy),
		CreatedAt:        m.CreatedAt,
	}
}
