package api

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Admin string
	Port  int
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Admin, validation.Required),
		validation.Field(&c.Port, validation.Required),
	)
}
