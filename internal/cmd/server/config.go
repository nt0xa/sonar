package server

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/russtone/sonar/internal/database"
)

type Config struct {
	DB     database.Config `json:"db"`
	Domain string          `json:"domain"`
	IP     string          `json:"ip"`

	DNS DNSConfig `json:"dns"`

	TLS TLSConfig `json:"tls"`

	Modules ModulesConfig `json:"modules"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.DB),
		validation.Field(&c.Domain, validation.Required, is.Domain),
		validation.Field(&c.IP, validation.Required, is.IP),
		validation.Field(&c.DNS),
		validation.Field(&c.TLS),
		validation.Field(&c.Modules),
	)
}
