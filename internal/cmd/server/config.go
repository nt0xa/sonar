package server

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	fsprov "github.com/knadh/koanf/providers/fs"
	"github.com/knadh/koanf/v2"
	"github.com/nt0xa/sonar/internal/utils"
	"github.com/nt0xa/sonar/internal/utils/valid"
)

var ConfigDefaults = map[string]any{
	"tls.type":                   "letsencrypt",
	"tls.letsencrypt.directory":  "./tls",
	"tls.letsencrypt.ca_dir_url": "https://acme-v02.api.letsencrypt.org/directory",
	"modules.enabled":            "api",
	"modules.api.port":           31337,
}

const ConfigFileName = "config.toml"

func LoadConfig(
	dir fs.FS,
	environFunc func() []string,
) (*Config, error) {
	k := koanf.New(".")

	// Load default values.
	if err := k.Load(confmap.Provider(ConfigDefaults, "."), nil); err != nil {
		return nil, fmt.Errorf("confmap failed: %w", err)
	}

	// Load config from TOML file.
	if dir != nil {
		if _, err := dir.Open(ConfigFileName); err == nil || !os.IsNotExist(err) {
			if err := k.Load(fsprov.Provider(dir, ConfigFileName), toml.Parser()); err != nil {
				return nil, fmt.Errorf("load from FS failed: %w", err)
			}
		}
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
		return nil, fmt.Errorf("load from env failed: %w", err)
	}

	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &cfg, nil
}

type Config struct {
	Domain    string
	IP        string
	GeoIP     GeoIPConfig
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
		validation.Field(&c.GeoIP),
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

//
// GeoIP
//

type GeoIPConfig struct {
	Enabled bool
	City    string
	ASN     string
}

func (c GeoIPConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Enabled),
		validation.Field(&c.City, validation.When(c.Enabled, validation.Required, validation.By(valid.File))),
		validation.Field(&c.ASN, validation.When(c.Enabled, validation.Required, validation.By(valid.File))),
	)
}
