package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type UsersDelete interface {
	UsersDelete(context.Context, UsersDeleteInput) (*UsersDeleteOutput, error)
}

type UsersDeleteInput struct {
	Name string
}

func (in UsersDeleteInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("name", in.Name, valid.Required),
	)
}

type UsersDeleteOutput User
