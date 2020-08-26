package actions

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

type PayloadsActions interface {
	PayloadsCreate(context.Context, PayloadsCreateParams) (PayloadsCreateResult, errors.Error)
	PayloadsUpdate(context.Context, PayloadsUpdateParams) (PayloadsUpdateResult, errors.Error)
	PayloadsDelete(context.Context, PayloadsDeleteParams) (PayloadsDeleteResult, errors.Error)
	PayloadsList(context.Context, PayloadsListParams) (PayloadsListResult, errors.Error)
}

type PayloadsHandler interface {
	PayloadsCreate(context.Context, PayloadsCreateResult)
	PayloadsList(context.Context, PayloadsListResult)
	PayloadsUpdate(context.Context, PayloadsUpdateResult)
	PayloadsDelete(context.Context, PayloadsDeleteResult)
}

type Payload struct {
	Subdomain       string    `json:"subdomain"`
	Name            string    `json:"name"`
	NotifyProtocols []string  `json:"notifyProtocols"`
	CreatedAt       time.Time `json:"createdAt"`
}

//
// Create
//

type PayloadsCreateParams struct {
	Name            string   `json:"name"`
	NotifyProtocols []string `json:"notifyProtocols"`
}

func (p PayloadsCreateParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required),
		validation.Field(&p.NotifyProtocols, validation.Each(validation.In(
			models.PayloadProtocolDNS,
			models.PayloadProtocolHTTP,
			models.PayloadProtocolSMTP,
		))),
	)
}

type PayloadsCreateResult *Payload

//
// Update
//

type PayloadsUpdateParams struct {
	Name            string   `json:"-"               path:"name"`
	NewName         string   `json:"name"`
	NotifyProtocols []string `json:"notifyProtocols"`
}

func (p PayloadsUpdateParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required),
		validation.Field(&p.NotifyProtocols, validation.Each(validation.In(
			models.PayloadProtocolDNS,
			models.PayloadProtocolHTTP,
			models.PayloadProtocolSMTP,
		))),
	)
}

type PayloadsUpdateResult *Payload

//
// Delete
//

type PayloadsDeleteParams struct {
	Name string `path:"name"`
}

func (p PayloadsDeleteParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

type PayloadsDeleteResult *Payload

//
// List
//

type PayloadsListParams struct {
	Name string `query:"name"`
}

func (p PayloadsListParams) Validate() error {
	return nil
}

type PayloadsListResult []*Payload
