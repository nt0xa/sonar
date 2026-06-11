package auditsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// PayloadsUpdate implements [service.Service].
func (s *Service) PayloadsUpdate(
	ctx context.Context,
	in service.PayloadsUpdateInput,
) (*service.PayloadsUpdateOutput, error) {
	out, err := s.Service.PayloadsUpdate(ctx, in)
	if err != nil {
		return out, err
	}

	s.writeAudit(
		ctx,
		database.AuditRecordActionTypeUpdate,
		database.AuditRecordResourceTypePayload,
		*out,
	)

	return out, nil
}
