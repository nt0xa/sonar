package database

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	DSN string `json:"dsn"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.DSN, validation.Required),
	)
}
