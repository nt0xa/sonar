package types

import "context"

type PayloadsList interface {
	PayloadsList(context.Context, PayloadsListInput) (PayloadsListOutput, error)
}

type PayloadsListInput struct {
	Name    string
	Page    uint
	PerPage uint
}

type PayloadsListOutput []Payload
