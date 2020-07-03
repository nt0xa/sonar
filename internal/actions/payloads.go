package actions

import (
	"database/sql"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils"
	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/bi-zone/sonar/internal/utils/slice"
)

type PayloadsActions interface {
	CreatePayload(*models.User, CreatePayloadParams) (CreatePayloadResult, errors.Error)
	UpdatePayload(*models.User, UpdatePayloadParams) (UpdatePayloadResult, errors.Error)
	DeletePayload(*models.User, DeletePayloadParams) (DeletePayloadResult, errors.Error)
	ListPayloads(*models.User, ListPayloadsParams) (ListPayloadsResult, errors.Error)
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

func (act *actions) CreatePayload(u *models.User, p CreatePayloadParams) (CreatePayloadResult, errors.Error) {

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	if _, err := act.db.PayloadsGetByUserAndName(u.ID, p.Name); err != sql.ErrNoRows {
		return nil, errors.Conflictf("payload with name %q already exist", p.Name)
	}

	subdomain, err := utils.GenerateRandomString(4)
	if err != nil {
		return nil, errors.Internal(err)
	}

	payload := &models.Payload{
		UserID:          u.ID,
		Subdomain:       subdomain,
		Name:            p.Name,
		NotifyProtocols: slice.StringsDedup(p.NotifyProtocols),
	}

	err = act.db.PayloadsCreate(payload)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return CreatePayloadResult(payload), nil
}

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

type UpdatePayloadResult = *MessageResult

func (act *actions) UpdatePayload(u *models.User, p UpdatePayloadParams) (UpdatePayloadResult, errors.Error) {

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(u.ID, p.Name)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.Name)
	} else if err != nil {
		return nil, errors.Internal(err)
	}

	if p.NewName != "" {
		payload.Name = p.NewName
	}

	if p.NotifyProtocols != nil {
		payload.NotifyProtocols = slice.StringsDedup(p.NotifyProtocols)
	}

	err = act.db.PayloadsUpdate(payload)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return &MessageResult{Message: "payload updated"}, nil
}

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

func (act *actions) DeletePayload(u *models.User, p DeletePayloadParams) (DeletePayloadResult, errors.Error) {

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(u.ID, p.Name)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("you don't have payload with name %q", p.Name)
	} else if err != nil {
		return nil, errors.Internal(err)
	}

	if err := act.db.PayloadsDelete(payload.ID); err != nil {
		return nil, errors.Internal(err)
	}

	return &MessageResult{Message: fmt.Sprintf("payload %q deleted", payload.Name)}, nil
}

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

func (act *actions) ListPayloads(u *models.User, p ListPayloadsParams) (ListPayloadsResult, errors.Error) {

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payloads, err := act.db.PayloadsFindByUserAndName(u.ID, p.Name)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return ListPayloadsResult(payloads), nil
}
