package actionsdb

import (
	"context"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/utils/errors"
)

func AuditRecord(m database.AuditRecord) actions.AuditRecord {
	return actions.AuditRecord{
		ID:           m.ID,
		UUID:         m.UUID,
		Action:       string(m.Action),
		ResourceType: string(m.ResourceType),
		Source:       string(m.Source),
		ActorID:      m.ActorID,
		ActorName:    m.ActorName,
		ActorMeta:    m.ActorMetadata,
		Resource:     m.Resource,
		CreatedAt:    m.CreatedAt,
	}
}

func (act *dbactions) AuditRecordsList(ctx context.Context, p actions.AuditRecordsListParams) (actions.AuditRecordsListResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}
	if u == nil || !u.IsAdmin {
		return nil, errors.Forbiddenf("admin only")
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	perPage := p.PerPage
	if perPage == 0 {
		perPage = 50
	}
	page := p.Page
	if page == 0 {
		page = 1
	}

	recs, err := act.db.AuditRecordsList(ctx, database.AuditRecordsListParams{
		ActorID:      p.ActorID,
		ActorName:    p.ActorName,
		ResourceType: p.ResourceType,
		Action:       p.Action,
		FromAt:       p.From,
		ToAt:         p.To,
		PageLimit:    int64(perPage),
		PageOffset:   int64((page - 1) * perPage),
	})
	if err != nil {
		return nil, errors.Internal(err)
	}

	res := make([]actions.AuditRecord, 0, len(recs))
	for _, r := range recs {
		res = append(res, AuditRecord(*r))
	}

	return res, nil
}

func (act *dbactions) AuditRecordsGet(ctx context.Context, p actions.AuditRecordsGetParams) (*actions.AuditRecordsGetResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}
	if u == nil || !u.IsAdmin {
		return nil, errors.Forbiddenf("admin only")
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	rec, err := act.db.AuditRecordsGetByID(ctx, p.ID)
	if err == database.ErrNoRows {
		return nil, errors.NotFoundf("audit with id %d not found", p.ID)
	} else if err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.AuditRecordsGetResult{AuditRecord: AuditRecord(*rec)}, nil
}
