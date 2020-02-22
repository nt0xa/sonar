package telegram

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Token string `json:"token"`
	Proxy string `json:"proxy"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Token, validation.Required),
		validation.Field(&c.Proxy),
	)
}
