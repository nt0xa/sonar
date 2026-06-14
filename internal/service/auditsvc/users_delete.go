package auditsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// UsersDelete implements [service.Service].
func (s *Service) UsersDelete(
	ctx context.Context,
	in service.UsersDeleteInput,
) (*service.UsersDeleteOutput, error) {
	out, err := s.ServerService.UsersDelete(ctx, in)
	if err != nil {
		return out, err
	}

	s.writeAudit(
		ctx,
		database.AuditRecordActionTypeDelete,
		database.AuditRecordResourceTypeUser,
		maskAPIToken((*service.User)(out)),
	)

	return out, nil
}
