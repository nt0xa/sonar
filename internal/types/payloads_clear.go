package types

import "context"

type PayloadsClear interface {
	PayloadsClear(context.Context, PayloadsClearInput) (PayloadsClearOutput, error)
}

type PayloadsClearInput struct {
	Name string
}

type PayloadsClearOutput = []Payload
