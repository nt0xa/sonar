package dbsvc

import (
	"context"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// AuditRecordsGet implements [service.Service].
func (s *Service) AuditRecordsGet(
	ctx context.Context,
	in service.AuditRecordsGetInput,
) (*service.AuditRecordsGetOutput, error) {
	if p := in.Validate(); p != nil {
		return nil, service.Validation(p)
	}

	if getUser(ctx) == nil {
		return nil, service.Unauthorized()
	}

	rec, err := s.db.AuditRecordsGetByID(ctx, in.ID)
	if errors.Is(err, database.ErrNoRows) {
		return nil, service.NotFoundf("audit record with id %d not found", in.ID)
	}
	if err != nil {
		return nil, err
	}

	return auditRecord(*rec), nil
}
