package actions

import (
	"database/sql"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/bi-zone/sonar/internal/utils/valid"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type DNSActions interface {
	CreateDNSRecord(*models.User, CreateDNSRecordParams) (CreateDNSRecordResult, errors.Error)
	DeleteDNSRecord(*models.User, DeleteDNSRecordParams) (DeleteDNSRecordResult, errors.Error)
	ListDNSRecords(*models.User, ListDNSRecordsParams) (ListDNSRecordsResult, errors.Error)
}

//
// Create
//

type CreateDNSRecordParams struct {
	PayloadName string   `json:"payloadName"`
	Name        string   `json:"name"`
	TTL         int      `json:"ttl"`
	Type        string   `json:"type"`
	Values      []string `json:"values"`
	Strategy    string   `json:"strategy"`
}

func (p CreateDNSRecordParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
		validation.Field(&p.Name, validation.Required, validation.By(valid.Subdomain)),
		validation.Field(&p.Type, valid.OneOf(models.DNSTypesAll)),
		validation.Field(&p.Values, validation.Required, validation.Each(valid.DNSRecord(p.Type))),
		validation.Field(&p.Strategy, valid.OneOf(models.DNSStrategiesAll)),
	)
}

type CreateDNSRecordResultData struct {
	Payload *models.Payload
	Record  *models.DNSRecord
}

type CreateDNSRecordResult = *CreateDNSRecordResultData

func (act *actions) CreateDNSRecord(u *models.User, p CreateDNSRecordParams) (CreateDNSRecordResult, errors.Error) {
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

	err = act.db.DNSRecordsCreate(rec)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return &CreateDNSRecordResultData{payload, rec}, nil
}

//
// Delete
//

type DeleteDNSRecordParams struct {
	PayloadName string `json:"payloadName"`
	Name        string `json:"name"`
	Type        string `json:"type"`
}

func (p DeleteDNSRecordParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
		validation.Field(&p.Name, validation.Required, validation.By(valid.Subdomain)),
		validation.Field(&p.Type, valid.OneOf(models.DNSTypesAll)),
	)
}

type DeleteDNSRecordResult = *MessageResult

func (act *actions) DeleteDNSRecord(u *models.User, p DeleteDNSRecordParams) (DeleteDNSRecordResult, errors.Error) {
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

	return &MessageResult{"record deleted"}, nil
}

//
// List
//

type ListDNSRecordsParams struct {
	PayloadName string `json:"payloadName"`
}

func (p ListDNSRecordsParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
	)
}

type ListDNSRecordsResultData struct {
	Payload *models.Payload
	Records []*models.DNSRecord
}

type ListDNSRecordsResult = *ListDNSRecordsResultData

func (act *actions) ListDNSRecords(u *models.User, p ListDNSRecordsParams) (ListDNSRecordsResult, errors.Error) {
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

	return &ListDNSRecordsResultData{payload, recs}, nil
}
