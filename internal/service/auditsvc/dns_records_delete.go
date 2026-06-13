package auditsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// DNSRecordsDelete implements [service.Service].
func (s *Service) DNSRecordsDelete(
	ctx context.Context,
	in service.DNSRecordsDeleteInput,
) (*service.DNSRecordsDeleteOutput, error) {
	out, err := s.ServerService.DNSRecordsDelete(ctx, in)
	if err != nil {
		return out, err
	}

	s.writeAudit(
		ctx,
		database.AuditRecordActionTypeDelete,
		database.AuditRecordResourceTypeDNSRecord,
		*out,
	)

	return out, nil
}
