package server

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/nt0xa/sonar/internal/utils/valid"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("tls.letsencrypt.directory", "./tls")
	viper.SetDefault("tls.letsencrypt.ca_dir_url", "https://acme-v02.api.letsencrypt.org/directory")
	viper.SetDefault("tls.letsencrypt.ca_insecure", false)

	viper.SetDefault("modules.enabled", "api")
	viper.SetDefault("modules.api.port", 31337)

	viper.SetDefault("modules.lark.tls_enabled", true)
}

type Config struct {
	Domain  string        `mapstructure:"domain"`
	IP      string        `mapstructure:"ip"`
	DB      DBConfig      `mapstructure:"db"`
	DNS     DNSConfig     `mapstructure:"dns"`
	TLS     TLSConfig     `mapstructure:"tls"`
	Modules ModulesConfig `mapstructure:"modules"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Domain, validation.Required, is.Domain),
		validation.Field(&c.IP, validation.Required, is.IP),
		validation.Field(&c.DB, validation.Required),
		validation.Field(&c.DNS),
		validation.Field(&c.TLS),
		validation.Field(&c.Modules),
	)
}

//
// DB
//

type DBConfig struct {
	DSN string `mapstructure:"dsn"`
}

func (c DBConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.DSN, validation.Required))
}

//
// DNS
//

type DNSConfig struct {
	Zone string `mapstructure:"zone"`
}

func (c DNSConfig) Validate() error {
	return validation.ValidateStruct(&c)
}

//
// TLS
//

type TLSConfig struct {
	Type        string               `mapstructure:"type"`
	Custom      TLSCustomConfig      `mapstructure:"custom"`
	LetsEncrypt TLSLetsEncryptConfig `mapstructure:"letsencrypt"`
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

// Custom

type TLSCustomConfig struct {
	Key  string `mapstructure:"key"`
	Cert string `mapstructure:"cert"`
}

func (c TLSCustomConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Key, validation.Required, validation.By(valid.File)),
		validation.Field(&c.Cert, validation.Required, validation.By(valid.File)),
	)
}

// LetsEncrypt

type TLSLetsEncryptConfig struct {
	Email      string `mapstructure:"email"`
	Directory  string `mapstructure:"directory"`
	CADirURL   string `mapstructure:"ca_dir_url"`
	CAInsecure bool   `mapstructure:"ca_insecure"`
}

func (c TLSLetsEncryptConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Email, validation.Required),
		validation.Field(&c.Directory, validation.Required, validation.By(valid.Directory)),
	)
}
