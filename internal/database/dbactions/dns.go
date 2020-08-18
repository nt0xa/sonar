package dbactions

import (
	"context"
	"database/sql"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func (act *dbactions) CreateDNSRecord(ctx context.Context, p actions.CreateDNSRecordParams) (actions.CreateDNSRecordResult, errors.Error) {
	u, e := actions.GetUser(ctx)
	if e != nil || u == nil {
		return nil, e
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(u.ID, p.PayloadName)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	}

	if _, err := act.db.DNSRecordsGetByPayloadNameType(payload.ID, p.Name, p.Type); err != sql.ErrNoRows {
		return nil, errors.Conflictf("dns records for payload %q with name %q and type %q already exist",
			p.PayloadName, p.Name, p.Type)
	}

	rec := &models.DNSRecord{
		PayloadID: payload.ID,
		Name:      p.Name,
		TTL:       p.TTL,
		Type:      p.Type,
		Values:    p.Values,
		Strategy:  p.Strategy,
	}

	if err := act.db.DNSRecordsCreate(rec); err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.CreateDNSRecordResultData{payload, rec}, nil
}

func (act *dbactions) DeleteDNSRecord(ctx context.Context, p actions.DeleteDNSRecordParams) (actions.DeleteDNSRecordResult, errors.Error) {
	u, e := actions.GetUser(ctx)
	if e != nil || u == nil {
		return nil, e
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(u.ID, p.PayloadName)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	}

	rec, err := act.db.DNSRecordsGetByPayloadNameType(payload.ID, p.Name, p.Type)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("dns records for payload %q with name %q and type %q not found",
			p.PayloadName, p.Name, p.Type)
	} else if err != nil {
		return nil, errors.Internal(err)
	}

	if err := act.db.DNSRecordsDelete(rec.ID); err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.MessageResult{"record deleted"}, nil
}

func (act *dbactions) ListDNSRecords(ctx context.Context, p actions.ListDNSRecordsParams) (actions.ListDNSRecordsResult, errors.Error) {
	u, e := actions.GetUser(ctx)
	if e != nil || u == nil {
		return nil, e
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(u.ID, p.PayloadName)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	}

	recs, err := act.db.DNSRecordsGetByPayloadID(payload.ID)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.ListDNSRecordsResultData{payload, recs}, nil
}
