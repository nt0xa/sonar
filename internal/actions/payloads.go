package actions

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

type PayloadsActions interface {
	CreatePayload(context.Context, CreatePayloadParams) (CreatePayloadResult, errors.Error)
	UpdatePayload(context.Context, UpdatePayloadParams) (UpdatePayloadResult, errors.Error)
	DeletePayload(context.Context, DeletePayloadParams) (DeletePayloadResult, errors.Error)
	ListPayloads(context.Context, ListPayloadsParams) (ListPayloadsResult, errors.Error)
}

//
// Create
//

type CreatePayloadParams struct {
	Name            string
	NotifyProtocols []string
}

func (p CreatePayloadParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required),
		validation.Field(&p.NotifyProtocols, validation.Each(validation.In(
			models.PayloadProtocolDNS,
			models.PayloadProtocolHTTP,
			models.PayloadProtocolSMTP,
		))),
	)
}

type CreatePayloadResult *models.Payload

//
// Update
//

type UpdatePayloadParams struct {
	Name            string
	NewName         string
	NotifyProtocols []string
}

func (p UpdatePayloadParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required),
		validation.Field(&p.NotifyProtocols, validation.Each(validation.In(
			models.PayloadProtocolDNS,
			models.PayloadProtocolHTTP,
			models.PayloadProtocolSMTP,
		))),
	)
}

type UpdatePayloadResult *models.Payload

//
// Delete
//

type DeletePayloadParams struct {
	Name string
}

func (p DeletePayloadParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

type DeletePayloadResult = *MessageResult

//
// List
//

type ListPayloadsParams struct {
	Name string
}

func (p ListPayloadsParams) Validate() error {
	return nil
}

type ListPayloadsResult []*models.Payload
