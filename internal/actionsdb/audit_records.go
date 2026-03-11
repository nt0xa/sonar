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
		ActorID:      m.Actor.ID,
		ActorName:    m.Actor.Name,
		ResourceType: m.Target.Type,
		ResourceID:   m.Target.ID,
		ResourceKey:  m.Target.Key,
		Action:       string(m.Operation),
		PayloadID:    m.Target.PayloadID,
		PayloadName:  m.Target.PayloadName,
		Meta:         m.Data,
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

	limit := p.Limit
	if limit == 0 {
		limit = 50
	}

	recs, err := act.db.AuditRecordsList(ctx, database.AuditRecordsListParams{
		ActorID:      p.ActorID,
		ActorName:    p.ActorName,
		ResourceType: p.ResourceType,
		ResourceID:   p.ResourceID,
		ResourceKey:  p.ResourceKey,
		Action:       p.Action,
		PayloadID:    p.PayloadID,
		PayloadName:  p.PayloadName,
		FromAt:       p.From,
		ToAt:         p.To,
		PageLimit:    int64(limit),
		PageOffset:   int64(p.Offset),
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
