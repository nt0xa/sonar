package auditsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// HTTPRoutesDelete implements [service.Service].
func (s *Service) HTTPRoutesDelete(
	ctx context.Context,
	in service.HTTPRoutesDeleteInput,
) (*service.HTTPRoutesDeleteOutput, error) {
	out, err := s.Service.HTTPRoutesDelete(ctx, in)
	if err != nil {
		return out, err
	}

	s.writeAudit(
		ctx,
		database.AuditRecordActionTypeDelete,
		database.AuditRecordResourceTypeHTTPRoute,
		*out,
	)

	return out, nil
}
