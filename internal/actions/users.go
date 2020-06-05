package actions

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/bi-zone/sonar/internal/utils/logger"
)

type CreateUserParams struct {
	Name   string
	Params database.UserParams
}

type CreateUserResult = *database.User

func (p CreateUserParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

func CreateUserAction(db *database.DB, log logger.StdLogger) *Action {
	return &Action{
		Params: &CreateUserParams{},
		Execute: func(u *database.User, params interface{}) (interface{}, error) {
			p, ok := params.(*CreateUserParams)
			if !ok {
				return nil, ErrParamsCast
			}

			user := &database.User{
				Name:   p.Name,
				Params: p.Params,
			}

			if err := db.UsersCreate(user); err != nil {
				return nil, errors.Internal(err)
			}

			return CreateUserResult(user), nil
		},
	}
}
