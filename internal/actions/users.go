package actions

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

type CreateUserParams struct {
	Name   string
	Params database.UserParams
}

func (p CreateUserParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

type CreateUserResult = *database.User

func (act *actions) CreateUser(p CreateUserParams) (CreateUserResult, error) {
	user := &database.User{
		Name:   p.Name,
		Params: p.Params,
	}

	if err := act.db.UsersCreate(user); err != nil {
		return nil, errors.Internal(err)
	}

	return CreateUserResult(user), nil
}
