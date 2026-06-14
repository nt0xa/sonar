package auditsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// PayloadsCreate implements [service.Service].
func (s *Service) PayloadsCreate(
	ctx context.Context,
	in service.PayloadsCreateInput,
) (*service.PayloadsCreateOutput, error) {
	out, err := s.ServerService.PayloadsCreate(ctx, in)
	if err != nil {
		return out, err
	}

	s.writeAudit(
		ctx,
		database.AuditRecordActionTypeCreate,
		database.AuditRecordResourceTypePayload,
		*out,
	)

	return out, nil
}
