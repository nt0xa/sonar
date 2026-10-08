package telegram

import (
	"github.com/nt0xa/sonar/pkg/valid"
)

type Config struct {
	Admin int64
	Token string
	Proxy string

	DisableEML bool `koanf:"disable_eml"`
}

func (c Config) Validate() valid.Problems {
	return valid.Validate(
		valid.Number("admin", c.Admin, valid.Required),
		valid.String("token", c.Token, valid.Required),
	)
}
