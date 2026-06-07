package service

import "context"

type ProfileGet interface {
	ProfileGet(context.Context ) (*ProfileGetOutput, error)
}

type ProfileGetOutput = User
