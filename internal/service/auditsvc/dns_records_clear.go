package auditsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// DNSRecordsClear implements [service.Service].
func (s *Service) DNSRecordsClear(
	ctx context.Context,
	in service.DNSRecordsClearInput,
) (service.DNSRecordsClearOutput, error) {
	out, err := s.Service.DNSRecordsClear(ctx, in)
	if err != nil {
		return out, err
	}

	for _, r := range out {
		s.writeAudit(
			ctx,
			database.AuditRecordActionTypeDelete,
			database.AuditRecordResourceTypeDNSRecord,
			r,
		)
	}

	return out, nil
}
