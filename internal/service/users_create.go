package service

import "context"

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

type UsersCreateOutput = User
