package slack

import (
	"github.com/nt0xa/sonar/pkg/valid"
)

type Config struct {
	Admin    string `koanf:"admin"`
	BotToken string `koanf:"bot_token"`
	AppToken string `koanf:"app_token"`
}

func (c Config) Validate() valid.Problems {
	return valid.Validate(
		valid.String("admin", c.Admin, valid.Required),
		valid.String("app_token", c.AppToken, valid.Required),
		valid.String("bot_token", c.BotToken, valid.Required),
	)
}
