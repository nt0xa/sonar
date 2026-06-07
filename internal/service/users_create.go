package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type UsersCreate interface {
	UsersCreate(context.Context, UsersCreateInput) (*UsersCreateOutput, error)
}

type UsersCreateInput struct {
	Name       string
	APIToken   *string
	TelegramID *int64
	LarkID     *string
	SlackID    *string
	IsAdmin    bool
}

func (in UsersCreateInput) Validate() v.Problems {
	return v.Struct(&in,
		v.String(&in.Name, v.Required),
	)
}

type UsersCreateOutput = User
