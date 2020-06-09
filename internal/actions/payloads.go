package actions

import (
	"database/sql"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

type PayloadsActions interface {
	CreatePayload(*database.User, CreatePayloadParams) (CreatePayloadResult, errors.Error)
	DeletePayload(*database.User, DeletePayloadParams) (DeletePayloadResult, errors.Error)
	ListPayloads(*database.User, ListPayloadsParams) (ListPayloadsResult, errors.Error)
}

//
// Create
//

type CreatePayloadParams struct {
	Name string
}

func (p CreatePayloadParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

type CreatePayloadResult = *database.Payload

func (act *actions) CreatePayload(u *database.User, p CreatePayloadParams) (CreatePayloadResult, errors.Error) {

	if _, err := act.db.PayloadsGetByUserAndName(u.ID, p.Name); err != sql.ErrNoRows {
		return nil, errors.Conflictf("you already have payload with name %q", p.Name)
	}

	subdomain, err := utils.GenerateRandomString(4)
	if err != nil {
		return nil, errors.Internal(err)
	}

	payload := &database.Payload{
		UserID:    u.ID,
		Subdomain: subdomain,
		Name:      p.Name,
	}

	err = act.db.PayloadsCreate(payload)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return CreatePayloadResult(payload), nil
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

func (act *actions) DeletePayload(u *database.User, p DeletePayloadParams) (DeletePayloadResult, errors.Error) {
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

type ListPayloadsResult = []*database.Payload

func (act *actions) ListPayloads(u *database.User, p ListPayloadsParams) (ListPayloadsResult, errors.Error) {

	payloads, err := act.db.PayloadsFindByUserAndName(u.ID, p.Name)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return ListPayloadsResult(payloads), nil
}
