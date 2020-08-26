package actions

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

type UsersActions interface {
	UsersCreate(context.Context, UsersCreateParams) (UsersCreateResult, errors.Error)
	UsersDelete(context.Context, UsersDeleteParams) (UsersDeleteResult, errors.Error)
}

type UsersHandler interface {
	UsersCreate(context.Context, UsersCreateResult)
	UsersDelete(context.Context, UsersDeleteResult)
}

type User struct {
	Name      string            `json:"name"`
	Params    models.UserParams `json:"params"`
	IsAdmin   bool              `json:"isAdmin"`
	CreatedAt time.Time         `json:"createdAt"`
}

//
// Create
//

type UsersCreateParams struct {
	Name    string            `json:"name"`
	Params  models.UserParams `json:"params"`
	IsAdmin bool              `json:"isAdmin"`
}

func (p UsersCreateParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

type UsersCreateResult *User

//
// Delete
//

type UsersDeleteParams struct {
	Name string `path:"name"`
}

func (p UsersDeleteParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

type UsersDeleteResult *User
