package auditsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// DNSRecordsCreate implements [service.Service].
func (s *Service) DNSRecordsCreate(
	ctx context.Context,
	in service.DNSRecordsCreateInput,
) (*service.DNSRecordsCreateOutput, error) {
	out, err := s.Service.DNSRecordsCreate(ctx, in)
	if err != nil {
		return out, err
	}

	s.writeAudit(
		ctx,
		database.AuditRecordActionTypeCreate,
		database.AuditRecordResourceTypeDNSRecord,
		*out,
	)

	return out, nil
}
