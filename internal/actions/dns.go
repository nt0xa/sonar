package actions

import (
	"context"
	"time"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/bi-zone/sonar/internal/utils/valid"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/miekg/dns"
)

type DNSActions interface {
	DNSRecordsCreate(context.Context, DNSRecordsCreateParams) (DNSRecordsCreateResult, errors.Error)
	DNSRecordsDelete(context.Context, DNSRecordsDeleteParams) (DNSRecordsDeleteResult, errors.Error)
	DNSRecordsList(context.Context, DNSRecordsListParams) (DNSRecordsListResult, errors.Error)
}

type DNSRecordsHandler interface {
	DNSRecordsCreate(context.Context, DNSRecordsCreateResult)
	DNSRecordsList(context.Context, DNSRecordsListResult)
	DNSRecordsDelete(context.Context, DNSRecordsDeleteResult)
}

type DNSRecord struct {
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	TTL       int       `json:"ttl"`
	Values    []string  `json:"values"`
	Strategy  string    `json:"strategy"`
	CreatedAt time.Time `json:"createdAt"`
}

func (r *DNSRecord) RRs(origin string) []dns.RR {
	rrs := make([]dns.RR, 0)
	for _, v := range r.Values {
		rrs = append(rrs, models.DNSStringToRR(v, r.Type, r.Name, origin, r.TTL))
	}
	return rrs
}

//
// Create
//

type DNSRecordsCreateParams struct {
	PayloadName string   `json:"payloadName"`
	Name        string   `json:"name"`
	TTL         int      `json:"ttl"`
	Type        string   `json:"type"`
	Values      []string `json:"values"`
	Strategy    string   `json:"strategy"`
}

func (p DNSRecordsCreateParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
		validation.Field(&p.Name, validation.Required, validation.By(valid.Subdomain)),
		validation.Field(&p.Type, valid.OneOf(models.DNSTypesAll)),
		validation.Field(&p.Values, validation.Required, validation.Each(valid.DNSRecord(p.Type))),
		validation.Field(&p.Strategy, valid.OneOf(models.DNSStrategiesAll)),
	)
}

type DNSRecordsCreateResultData struct {
	Payload *Payload   `json:"payload"`
	Record  *DNSRecord `json:"record"`
}

type DNSRecordsCreateResult = *DNSRecordsCreateResultData

//
// Delete
//

type DNSRecordsDeleteParams struct {
	PayloadName string `path:"payloadName"`
	Name        string `path:"name"`
	Type        string `path:"type"`
}

func (p DNSRecordsDeleteParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
		validation.Field(&p.Name, validation.Required, validation.By(valid.Subdomain)),
		validation.Field(&p.Type, valid.OneOf(models.DNSTypesAll)),
	)
}

type DNSRecordsDeleteResultData struct {
	Payload *Payload   `json:"payload"`
	Record  *DNSRecord `json:"record"`
}

type DNSRecordsDeleteResult = *DNSRecordsDeleteResultData

//
// List
//

type DNSRecordsListParams struct {
	PayloadName string `path:"payloadName"`
}

func (p DNSRecordsListParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
	)
}

type DNSRecordsListResultData struct {
	Payload *Payload     `json:"payload"`
	Records []*DNSRecord `json:"records"`
}

type DNSRecordsListResult = *DNSRecordsListResultData
