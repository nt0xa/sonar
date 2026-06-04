package service

import "context"

type PayloadsDelete interface {
	PayloadsDelete(context.Context, PayloadsDeleteInput) (*PayloadsDeleteOutput, error)
}

type PayloadsDeleteInput struct {
	Name string
}

type PayloadsDeleteOutput = Payload
