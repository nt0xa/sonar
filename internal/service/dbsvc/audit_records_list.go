package dbsvc

import (
	"cmp"
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// AuditRecordsList implements [service.Service].
func (s *Service) AuditRecordsList(
	ctx context.Context,
	in service.AuditRecordsListInput,
) (service.AuditRecordsListOutput, error) {
	if _, ok := service.GetUserID(ctx); !ok {
		return nil, service.Unauthorized()
	}

	perPage := cmp.Or(in.PerPage, 10)
	page := cmp.Or(in.Page, 1)

	recs, err := s.db.AuditRecordsList(ctx, database.AuditRecordsListParams{
		ActorID:      in.ActorID,
		ActorName:    in.ActorName,
		ResourceType: string(in.ResourceType),
		Action:       string(in.Action),
		FromAt:       in.From,
		ToAt:         in.To,
		PageOffset:   int64((page - 1) * perPage),
		PageLimit:    int64(perPage),
	})
	if err != nil {
		return nil, err
	}

	out := make([]service.AuditRecord, len(recs))

	for i, r := range recs {
		out[i] = *auditRecord(*r)
	}

	return out, nil
}
