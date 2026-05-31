package types

import "context"

type ProfileGet interface {
	ProfileGet(context.Context, ProfileGetInput) (*ProfileGetOutput, error)
}

type ProfileGetInput struct{}

type ProfileGetOutput User
