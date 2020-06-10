package actions

import (
	"database/sql"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

type UsersActions interface {
	CreateUser(*database.User, CreateUserParams) (CreateUserResult, errors.Error)
	DeleteUser(*database.User, DeleteUserParams) (DeleteUserResult, errors.Error)
}

type CreateUserParams struct {
	Name   string
	Params database.UserParams
}

func (p CreateUserParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

type CreateUserResult = *database.User

func (act *actions) CreateUser(u *database.User, p CreateUserParams) (CreateUserResult, errors.Error) {

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	if _, err := act.db.UsersGetByName(p.Name); err != sql.ErrNoRows {
		return nil, errors.Conflictf("user with name %q already exist", p.Name)
	}

	user := &database.User{
		Name:   p.Name,
		Params: p.Params,
	}

	if err := act.db.UsersCreate(user); err != nil {
		return nil, errors.Internal(err)
	}

	return CreateUserResult(user), nil
}

type DeleteUserParams struct {
	Name string
}

func (p DeleteUserParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

type DeleteUserResult = *MessageResult

func (act *actions) DeleteUser(u *database.User, p DeleteUserParams) (DeleteUserResult, errors.Error) {

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	user := &database.User{
		Name: p.Name,
	}

	if err := act.db.UsersDelete(user.ID); err != nil {
		return nil, errors.Internal(err)
	}

	return &MessageResult{Message: fmt.Sprintf("user %q deleted", user.Name)}, nil
}
