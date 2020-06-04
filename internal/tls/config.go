package tls

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/bi-zone/sonar/internal/utils/valid"
)

type Config struct {
	Type        string            `json:"type"`
	Custom      CustomConfig      `json:"custom"`
	LetsEncrypt LetsEncryptConfig `json:"letsencrypt"`
}

type CustomConfig struct {
	Key  string `json:"key"`
	Cert string `json:"cert"`
}

type LetsEncryptConfig struct {
	Email      string `json:"email"`
	Directory  string `json:"directory"`
	CADirURL   string `json:"caDirUrl" default:"https://acme-v02.api.letsencrypt.org/directory"`
	CAInsecure bool   `json:"caInsecure"`
}

func (c Config) Validate() error {
	rules := make([]*validation.FieldRules, 0)

	rules = append(rules,
		validation.Field(&c.Type, validation.Required, validation.In("custom", "letsencrypt")))

	switch c.Type {
	case "custom":
		rules = append(rules, validation.Field(&c.Custom))
	case "letsencrypt":
		rules = append(rules, validation.Field(&c.LetsEncrypt))
	}

	return validation.ValidateStruct(&c, rules...)
}

func (c CustomConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Key, validation.Required, validation.By(valid.File)),
		validation.Field(&c.Cert, validation.Required, validation.By(valid.File)),
	)
}

func (c LetsEncryptConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Email, validation.Required),
		validation.Field(&c.Directory, validation.Required, validation.By(valid.Directory)),
	)
}
