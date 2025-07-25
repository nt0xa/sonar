package actionsdb

import (
	"context"
	"database/sql"
	"strings"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/utils/errors"
)

func DNSRecord(m models.DNSRecord, payloadSubdomain string) actions.DNSRecord {
	return actions.DNSRecord{
		Index:            m.Index,
		PayloadSubdomain: payloadSubdomain,
		Name:             m.Name,
		Type:             m.Type,
		TTL:              m.TTL,
		Values:           m.Values,
		Strategy:         m.Strategy,
		CreatedAt:        m.CreatedAt,
	}
}

func (act *dbactions) DNSRecordsCreate(ctx context.Context, p actions.DNSRecordsCreateParams) (*actions.DNSRecordsCreateResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(ctx, u.ID, p.PayloadName)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	} else if err != nil {
		return nil, errors.Internal(err)
	}

	if _, err := act.db.DNSRecordsGetByPayloadNameAndType(ctx, payload.ID, p.Name, strings.ToUpper(p.Type)); err != sql.ErrNoRows {
		return nil, errors.Conflictf("dns records for payload %q with name %q and type %q already exist",
			p.PayloadName, p.Name, p.Type)
	}

	rec := &models.DNSRecord{
		PayloadID: payload.ID,
		Name:      p.Name,
		TTL:       p.TTL,
		Type:      strings.ToUpper(p.Type),
		Values:    p.Values,
		Strategy:  p.Strategy,
	}

	if err := act.db.DNSRecordsCreate(ctx, rec); err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.DNSRecordsCreateResult{DNSRecord: DNSRecord(*rec, payload.Subdomain)}, nil
}

func (act *dbactions) DNSRecordsDelete(ctx context.Context, p actions.DNSRecordsDeleteParams) (*actions.DNSRecordsDeleteResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(ctx, u.ID, p.PayloadName)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	}

	rec, err := act.db.DNSRecordsGetByPayloadIDAndIndex(ctx, payload.ID, p.Index)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("dns record for payload %q with index %d not found",
			p.PayloadName, p.Index)
	} else if err != nil {
		return nil, errors.Internal(err)
	}

	if err := act.db.DNSRecordsDelete(ctx, rec.ID); err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.DNSRecordsDeleteResult{DNSRecord: DNSRecord(*rec, payload.Subdomain)}, nil
}

func (act *dbactions) DNSRecordsClear(ctx context.Context, p actions.DNSRecordsClearParams) (actions.DNSRecordsClearResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(ctx, u.ID, p.PayloadName)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	}

	var recs []*models.DNSRecord

	if p.Name != "" {
		recs, err = act.db.DNSRecordsDeleteAllByPayloadIDAndName(ctx, payload.ID, p.Name)
	} else {
		recs, err = act.db.DNSRecordsDeleteAllByPayloadID(ctx, payload.ID)
	}
	if err != nil {
		return nil, errors.Internal(err)
	}

	res := make([]actions.DNSRecord, 0)

	for _, r := range recs {
		res = append(res, DNSRecord(*r, payload.Subdomain))
	}

	return res, nil
}

func (act *dbactions) DNSRecordsList(ctx context.Context, p actions.DNSRecordsListParams) (actions.DNSRecordsListResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(ctx, u.ID, p.PayloadName)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	}

	recs, err := act.db.DNSRecordsGetByPayloadID(ctx, payload.ID)
	if err != nil {
		return nil, errors.Internal(err)
	}

	res := make([]actions.DNSRecord, 0)

	for _, r := range recs {
		res = append(res, DNSRecord(*r, payload.Subdomain))
	}

	return res, nil
}
