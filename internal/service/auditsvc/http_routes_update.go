package auditsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// HTTPRoutesUpdate implements [service.Service].
func (s *Service) HTTPRoutesUpdate(
	ctx context.Context,
	in service.HTTPRoutesUpdateInput,
) (*service.HTTPRoutesUpdateOutput, error) {
	out, err := s.ServerService.HTTPRoutesUpdate(ctx, in)
	if err != nil {
		return out, err
	}

	s.writeAudit(
		ctx,
		database.AuditRecordActionTypeUpdate,
		database.AuditRecordResourceTypeHTTPRoute,
		*out,
	)

	return out, nil
}
