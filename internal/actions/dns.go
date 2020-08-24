package actions

import (
	"context"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/bi-zone/sonar/internal/utils/valid"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type DNSActions interface {
	CreateDNSRecord(context.Context, CreateDNSRecordParams) (CreateDNSRecordResult, errors.Error)
	DeleteDNSRecord(context.Context, DeleteDNSRecordParams) (DeleteDNSRecordResult, errors.Error)
	ListDNSRecords(context.Context, ListDNSRecordsParams) (ListDNSRecordsResult, errors.Error)
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
