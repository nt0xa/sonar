package actions

import (
	"context"

	"github.com/bi-zone/sonar/internal/utils/errors"
)

type UserActions interface {
	UserCurrent(context.Context) (UserCurrentResult, errors.Error)
}

type UserHandler interface {
	UserCurrent(context.Context, UserCurrentResult)
}

type UserCurrentResult *User
