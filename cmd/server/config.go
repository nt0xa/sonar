package main

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/modules"
	"github.com/bi-zone/sonar/internal/tls"
)

type Config struct {
	DB     database.Config `json:"db"`
	Domain string          `json:"domain"`
	IP     string          `json:"ip"`

	TLS tls.Config `json:"tls"`

	Modules modules.Config `json:"modules"`
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
