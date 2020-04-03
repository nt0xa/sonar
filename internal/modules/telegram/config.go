package telegram

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Admin  int64  `json:"admin"`
	Token  string `json:"token"`
	Proxy  string `json:"proxy"`
	Domain string `json:"domain" envconfig:"DOMAIN"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Admin, validation.Required),
		validation.Field(&c.Token, validation.Required),
		validation.Field(&c.Proxy),
	)
}
