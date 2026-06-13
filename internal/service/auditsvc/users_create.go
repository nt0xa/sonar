package auditsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// UsersCreate implements [service.Service].
func (s *Service) UsersCreate(
	ctx context.Context,
	in service.UsersCreateInput,
) (*service.UsersCreateOutput, error) {
	out, err := s.ServerService.UsersCreate(ctx, in)
	if err != nil {
		return out, err
	}

	s.writeAudit(
		ctx,
		database.AuditRecordActionTypeCreate,
		database.AuditRecordResourceTypeUser,
		maskAPIToken(out),
	)

	return out, nil
}
