package server

import (
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/nt0xa/sonar/internal/utils"
	"github.com/nt0xa/sonar/internal/utils/valid"
)

func GetConfig(
	defaults map[string]any,
	configData []byte,
	environFunc func() []string,
) (*Config, error) {
	k := koanf.New(".")

	// Load default values.
	if err := k.Load(confmap.Provider(defaults, "."), nil); err != nil {
		return nil, fmt.Errorf("load config from confmap: %w", err)
	}

	// Load config from TOML file.
	if err := k.Load(rawbytes.Provider(configData), toml.Parser()); err != nil {
		return nil, fmt.Errorf("load config from rawbytes: %w", err)
	}
	var cfg Config

	cfgKeys := make(map[string]string)
	for _, key := range utils.StructKeys(cfg, "koanf") {
		envKey := "SONAR_" + strings.ReplaceAll(strings.ToUpper(key), ".", "_")
		cfgKeys[envKey] = key
	}

	// Load config from environment variables.
	if err := k.Load(env.Provider(".", env.Opt{
		Prefix: "SONAR_",
		TransformFunc: func(k, v string) (string, any) {
			key, ok := cfgKeys[k]
			if !ok {
				return "", ""
			}
			if strings.Contains(v, ",") {
				return key, strings.Split(v, ",")
			}
			return key, v
		},
		EnvironFunc: environFunc,
	}), nil); err != nil {
		return nil, fmt.Errorf("load config from env: %w", err)
	}

	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

var ConfigDefaults = map[string]any{
	"tls.type":                   "letsencrypt",
	"tls.letsencrypt.directory":  "./tls",
	"tls.letsencrypt.ca_dir_url": "https://acme-v02.api.letsencrypt.org/directory",
	"modules.enabled":            "api",
	"modules.api.port":           31337,
}

type Config struct {
	Domain    string
	IP        string
	DB        DBConfig
	DNS       DNSConfig
	TLS       TLSConfig
	Telemetry TelemetryConfig
	Modules   ModulesConfig
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
// Telemetry
//

type TelemetryConfig struct {
	Enabled bool
}

func (c TelemetryConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Enabled),
	)
}

//
// DB
//

type DBConfig struct {
	DSN string
}

func (c DBConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.DSN, validation.Required))
}

//
// DNS
//

type DNSConfig struct {
	Zone string
}

func (c DNSConfig) Validate() error {
	return validation.ValidateStruct(&c)
}

//
// TLS
//

type TLSConfig struct {
	Type        string
	Custom      TLSCustomConfig
	LetsEncrypt TLSLetsEncryptConfig
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
	Key  string
	Cert string
}

func (c TLSCustomConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Key, validation.Required, validation.By(valid.File)),
		validation.Field(&c.Cert, validation.Required, validation.By(valid.File)),
	)
}

// LetsEncrypt

type TLSLetsEncryptConfig struct {
	Email      string
	Directory  string
	CADirURL   string `koanf:"ca_dir_url"`
	CAInsecure bool   `koanf:"ca_insecure"`
}

func (c TLSLetsEncryptConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Email, validation.Required),
		validation.Field(&c.Directory, validation.Required, validation.By(valid.Directory)),
	)
}
