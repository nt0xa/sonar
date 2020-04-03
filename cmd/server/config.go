package main

import (
	"errors"
	"os"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/modules"
)

type Config struct {
	DB     database.Config `json:"db"`
	Domain string          `json:"domain"`
	IP     string          `json:"ip"`

	TLS TLSConfig `json:"tls"`

	Modules modules.Config `json:"modules"`
}

type TLSConfig struct {
	Type        string            `json:"type"`
	Custom      CustomConfig      `json:"custom"`
	LetsEncrypt LetsEncryptConfig `json:"letsencrypt"`
}

type CustomConfig struct {
	Key  string `json:"key"`
	Cert string `json:"cert"`
}

type LetsEncryptConfig struct {
	Email     string `json:"email"`
	Directory string `json:"directory"`
}

func validateFile(value interface{}) error {
	path, _ := value.(string)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	return nil
}

func validateDirectory(value interface{}) error {
	path, _ := value.(string)

	if fi, err := os.Stat(path); os.IsNotExist(err) {
		return err
	} else if fi.Mode().IsRegular() {
		return errors.New("must be directory")
	}

	return nil
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.DB),
		validation.Field(&c.Domain, validation.Required, is.Domain),
		validation.Field(&c.IP, validation.Required, is.IPv4),
		validation.Field(&c.TLS),
		validation.Field(&c.Modules),
	)
}

func (c TLSConfig) Validate() error {
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
		validation.Field(&c.Key, validation.Required, validation.By(validateFile)),
		validation.Field(&c.Cert, validation.Required, validation.By(validateFile)),
	)
}

func (c LetsEncryptConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Email, validation.Required),
		validation.Field(&c.Directory, validation.Required, validation.By(validateDirectory)),
	)
}
