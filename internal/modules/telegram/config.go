package telegram

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Admin int64  `mapstructure:"admin"`
	Token string `mapstructure:"token"`
	Proxy string `mapstructure:"proxy"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Admin, validation.Required),
		validation.Field(&c.Token, validation.Required),
		validation.Field(&c.Proxy),
	)
}
