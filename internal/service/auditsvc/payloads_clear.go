package auditsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// PayloadsClear implements [service.Service].
func (s *Service) PayloadsClear(
	ctx context.Context,
	in service.PayloadsClearInput,
) (service.PayloadsClearOutput, error) {
	out, err := s.ServerService.PayloadsClear(ctx, in)
	if err != nil {
		return out, err
	}

	for _, p := range out {
		s.writeAudit(
			ctx,
			database.AuditRecordActionTypeDelete,
			database.AuditRecordResourceTypePayload,
			p,
		)
	}

	return out, nil
}
