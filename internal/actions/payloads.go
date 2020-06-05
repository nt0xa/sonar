package actions

import (
	"database/sql"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils"
	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/bi-zone/sonar/internal/utils/logger"
)

type CreatePayloadParams struct {
	Name string
}

type CreatePayloadResult = *database.Payload

func (p CreatePayloadParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

func CreatePayloadAction(db *database.DB, log logger.StdLogger) *Action {
	return &Action{
		Params: &CreatePayloadParams{},
		Execute: func(u *database.User, params interface{}) (interface{}, error) {
			var p *CreatePayloadParams

			p, ok := params.(*CreatePayloadParams)
			if !ok {
				return nil, ErrParamsCast
			}

			if _, err := db.PayloadsGetByUserAndName(u.ID, p.Name); err != sql.ErrNoRows {
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

			err = db.PayloadsCreate(payload)
			if err != nil {
				return nil, errors.Internal(err)
			}

			return CreatePayloadResult(payload), nil
		},
	}
}

type DeletePayloadParams struct {
	Name string
}

type DeletePayloadResult = *MessageResult

func (p DeletePayloadParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

func DeletePayloadAction(db *database.DB, log logger.StdLogger) *Action {
	return &Action{
		Params: &DeletePayloadParams{},
		Execute: func(u *database.User, params interface{}) (interface{}, error) {
			var p *DeletePayloadParams

			p, ok := params.(*DeletePayloadParams)
			if !ok {
				return nil, ErrParamsCast
			}

			payload, err := db.PayloadsGetByUserAndName(u.ID, p.Name)
			if err == sql.ErrNoRows {
				return nil, errors.NotFoundf("you don't have payload with name %q", p.Name)
			} else if err != nil {
				return nil, errors.Internal(err)
			}

			if err := db.PayloadsDelete(payload.ID); err != nil {
				return nil, errors.Internal(err)
			}

			return &MessageResult{Message: "payload deleted"}, nil
		},
	}
}

type ListPayloadsParams struct {
	Name string
}

type ListPayloadsResult = []*database.Payload

func (p ListPayloadsParams) Validate() error {
	return nil
}

func ListPayloadsAction(db *database.DB, log logger.StdLogger) *Action {
	return &Action{
		Params: &ListPayloadsParams{},
		Execute: func(u *database.User, params interface{}) (interface{}, error) {
			var p *ListPayloadsParams

			p, ok := params.(*ListPayloadsParams)
			if !ok {
				return nil, ErrParamsCast
			}

			payloads, err := db.PayloadsFindByUserAndName(u.ID, p.Name)
			if err != nil {
				return nil, errors.Internal(err)
			}

			return ListPayloadsResult(payloads), nil
		},
	}
}
