package auditsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// HTTPRoutesClear implements [service.Service].
func (s *Service) HTTPRoutesClear(
	ctx context.Context,
	in service.HTTPRoutesClearInput,
) (service.HTTPRoutesClearOutput, error) {
	out, err := s.ServerService.HTTPRoutesClear(ctx, in)
	if err != nil {
		return out, err
	}

	for _, r := range out {
		s.writeAudit(
			ctx,
			database.AuditRecordActionTypeDelete,
			database.AuditRecordResourceTypeHTTPRoute,
			r,
		)
	}

	return out, nil
}
