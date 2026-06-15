package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
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

func (in UsersCreateInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("name", in.Name, valid.Required),
	)
}

type UsersCreateOutput User
