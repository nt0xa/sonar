package auditsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// PayloadsDelete implements [service.Service].
func (s *Service) PayloadsDelete(
	ctx context.Context,
	in service.PayloadsDeleteInput,
) (*service.PayloadsDeleteOutput, error) {
	out, err := s.ServerService.PayloadsDelete(ctx, in)
	if err != nil {
		return out, err
	}

	s.writeAudit(
		ctx,
		database.AuditRecordActionTypeDelete,
		database.AuditRecordResourceTypePayload,
		*out,
	)

	return out, nil
}
