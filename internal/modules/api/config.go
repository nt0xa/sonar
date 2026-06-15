package api

import (
	"github.com/nt0xa/sonar/pkg/valid"
)

type Config struct {
	Admin string
	Port  int
}

func (c Config) Validate() valid.Problems {
	return valid.Validate(
		valid.String("admin", c.Admin, valid.Required),
		valid.Number("port", c.Port, valid.Required, valid.Min(1), valid.Max(65535)),
	)
}
