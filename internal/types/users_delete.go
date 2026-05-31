package types

import "context"

type UsersDelete interface {
	UsersDelete(context.Context, UsersDeleteInput) (*UsersDeleteOutput, error)
}

type UsersDeleteInput struct {
	Name string
}

type UsersDeleteOutput User
