package auditsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// HTTPRoutesCreate implements [service.Service].
func (s *Service) HTTPRoutesCreate(
	ctx context.Context,
	in service.HTTPRoutesCreateInput,
) (*service.HTTPRoutesCreateOutput, error) {
	out, err := s.ServerService.HTTPRoutesCreate(ctx, in)
	if err != nil {
		return out, err
	}

	s.writeAudit(
		ctx,
		database.AuditRecordActionTypeCreate,
		database.AuditRecordResourceTypeHTTPRoute,
		*out,
	)

	return out, nil
}
