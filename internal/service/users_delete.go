package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type UsersDelete interface {
	UsersDelete(context.Context, UsersDeleteInput) (*UsersDeleteOutput, error)
}

type UsersDeleteInput struct {
	Name string
}

func (in UsersDeleteInput) Validate() v.Problems {
	return v.Struct(
		v.String("name", in.Name).Required(),
	)
}

type UsersDeleteOutput = User
