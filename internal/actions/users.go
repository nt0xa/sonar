package actions

import (
	"database/sql"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

type UsersActions interface {
	CreateUser(*models.User, CreateUserParams) (CreateUserResult, errors.Error)
	DeleteUser(*models.User, DeleteUserParams) (DeleteUserResult, errors.Error)
}

type CreateUserParams struct {
	Name      string
	Params    models.UserParams
	IsAdmin   bool
	CreatedBy *int64
}

func (p CreateUserParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

type CreateUserResult = *models.User

func (act *actions) CreateUser(u *models.User, p CreateUserParams) (CreateUserResult, errors.Error) {

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	if _, err := act.db.UsersGetByName(p.Name); err != sql.ErrNoRows {
		return nil, errors.Conflictf("user with name %q already exist", p.Name)
	}

	user := &models.User{
		Name:      p.Name,
		Params:    p.Params,
		IsAdmin:   p.IsAdmin,
		CreatedBy: p.CreatedBy,
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

func (act *actions) DeleteUser(u *models.User, p DeleteUserParams) (DeleteUserResult, errors.Error) {

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	user, err := act.db.UsersGetByName(p.Name)
	if err != nil {
		return nil, errors.NotFoundf("user with name %q not found", p.Name)
	}

	if err := act.db.UsersDelete(user.ID); err != nil {
		return nil, errors.Internal(err)
	}

	return &MessageResult{Message: fmt.Sprintf("user %q deleted", user.Name)}, nil
}
