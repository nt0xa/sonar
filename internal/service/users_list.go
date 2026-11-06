package service

import "context"

type UsersList interface {
	UsersList(context.Context, UsersListInput) (UsersListOutput, error)
}

type UsersListInput struct {
	Name    string
	Page    uint
	PerPage uint
}

type UsersListOutput []User
