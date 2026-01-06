package slack

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Admin    string `koanf:"admin"`
	BotToken string `koanf:"bot_token"`
	AppToken string `koanf:"app_token"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Admin, validation.Required),
		validation.Field(&c.AppToken, validation.Required),
		validation.Field(&c.BotToken, validation.Required),
	)
}
