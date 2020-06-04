package actions

import (
	"database/sql"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils"
	"github.com/bi-zone/sonar/internal/utils/logger"
)

type NewPayloadParams struct {
	Name string
}

func (p *NewPayloadParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

func NewPayloadAction(db *database.DB, log logger.StdLogger) *Action {
	return &Action{
		Params: &NewPayloadParams{},
		Execute: func(u *database.User, params interface{}) (*ActionResult, error) {
			var p *NewPayloadParams

			p, ok := params.(*NewPayloadParams)
			if !ok {
				return nil, castError
			}

			if _, err := db.PayloadsGetByUserAndName(u.ID, p.Name); err != sql.ErrNoRows {
				return nil, ErrConflict("you already have payload with name %q", p.Name)
			}

			subdomain, err := utils.GenerateRandomString(4)
			if err != nil {
				return nil, ErrInternal(err)
			}

			payload := &database.Payload{
				UserID:    u.ID,
				Subdomain: subdomain,
				Name:      p.Name,
			}

			err = db.PayloadsCreate(payload)
			if err != nil {
				return nil, ErrInternal(err)
			}

			return &ActionResult{Data: payload}, nil
		},
	}
}

type DeletePayloadParams struct {
	Name string
}

func (p *DeletePayloadParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

func DeletePayloadAction(db *database.DB, log logger.StdLogger) *Action {
	return &Action{
		Params: &DeletePayloadParams{},
		Execute: func(u *database.User, params interface{}) (*ActionResult, error) {
			var p *DeletePayloadParams

			payload, err := db.PayloadsGetByUserAndName(u.ID, p.Name)
			if err == sql.ErrNoRows {
				return nil, ErrNotFound("you don't have payload with name %q", p.Name)
			} else if err != nil {
				return nil, ErrInternal(err)
			}

			if err := db.PayloadsDelete(payload.ID); err != nil {
				return nil, ErrInternal(err)
			}

			return &ActionResult{Message: "payload deleted"}, nil
		},
	}
}

type ListPayloadsParams struct {
	PartName string
}

func (p *ListPayloadsParams) Validate() error {
	return nil
}

func ListPayloadsAction(db *database.DB, log logger.StdLogger) *Action {
	return &Action{
		Params: &ListPayloadsParams{},
		Execute: func(u *database.User, params interface{}) (*ActionResult, error) {
			var p *ListPayloadsParams

			payloads, err := db.PayloadsFindByUserAndName(u.ID, p.PartName)
			if err != nil {
				return nil, ErrInternal(err)
			}

			return &ActionResult{Data: payloads}, nil
		},
	}
}
