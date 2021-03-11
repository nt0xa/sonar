package server

import (
	"github.com/bi-zone/sonar/internal/database"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Config struct {
	DB     database.Config `json:"db"`
	Domain string          `json:"domain"`
	IP     string          `json:"ip"`

	TLS TLSConfig `json:"tls"`

	Modules ModulesConfig `json:"modules"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.DB),
		validation.Field(&c.Domain, validation.Required, is.Domain),
		validation.Field(&c.IP, validation.Required, is.IP),
		validation.Field(&c.TLS),
		validation.Field(&c.Modules),
	)
}
