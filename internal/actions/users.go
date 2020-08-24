package actions

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

type UsersActions interface {
	CreateUser(context.Context, CreateUserParams) (CreateUserResult, errors.Error)
	DeleteUser(context.Context, DeleteUserParams) (DeleteUserResult, errors.Error)
}

type CreateUserParams struct {
	Name    string
	Params  models.UserParams
	IsAdmin bool
}

func (p CreateUserParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

type CreateUserResult *models.User

type DeleteUserParams struct {
	Name string
}

func (p DeleteUserParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

type DeleteUserResult = *MessageResult
