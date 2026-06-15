package server

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	fsprov "github.com/knadh/koanf/providers/fs"
	"github.com/knadh/koanf/v2"
	"github.com/nt0xa/sonar/internal/utils"
	"github.com/nt0xa/sonar/pkg/valid"
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

	if p := cfg.Validate(); !p.Ok() {
		return nil, fmt.Errorf("validation failed: %w", p)
	}

	return &cfg, nil
}

type Config struct {
	Domain    string
	IP        string
	GeoIP     GeoIPConfig
	DB        DBConfig
	Audit     AuditConfig
	DNS       DNSConfig
	TLS       TLSConfig
	Telemetry TelemetryConfig
	Modules   ModulesConfig
}

func (c Config) Validate() valid.Problems {
	return valid.Validate(
		valid.String("domain", c.Domain, valid.Required, valid.Domain),
		valid.String("ip", c.IP, valid.Required, valid.IP),
		valid.Struct("db", c.DB),
		valid.Struct("geoip", c.GeoIP),
		valid.Struct("tls", c.TLS),
		valid.Struct("modules", c.Modules),
	)
}

type AuditConfig struct {
	Enabled bool
}

//
// Telemetry
//

type TelemetryConfig struct {
	Enabled bool
}

//
// DB
//

type DBConfig struct {
	DSN string
}

func (c DBConfig) Validate() valid.Problems {
	return valid.Validate(
		valid.String("dsn", c.DSN, valid.Required),
	)
}

//
// DNS
//

type DNSConfig struct {
	Zone string
}

//
// TLS
//

type TLSConfig struct {
	Type        string
	Custom      TLSCustomConfig
	LetsEncrypt TLSLetsEncryptConfig
}

func (c TLSConfig) Validate() valid.Problems {
	fields := []valid.Validatable{
		valid.String("type", c.Type, valid.Required, valid.In("custom", "letsencrypt")),
	}

	switch c.Type {
	case "custom":
		fields = append(fields, valid.Struct("custom", c.Custom))
	case "letsencrypt":
		fields = append(fields, valid.Struct("letsencrypt", c.LetsEncrypt))
	}

	return valid.Validate(fields...)
}

// Custom

type TLSCustomConfig struct {
	Key  string
	Cert string
}

func (c TLSCustomConfig) Validate() valid.Problems {
	return valid.Validate(
		valid.String("key", c.Key, valid.Required, file),
		valid.String("cert", c.Cert, valid.Required, file),
	)
}

// LetsEncrypt

type TLSLetsEncryptConfig struct {
	Email      string
	Directory  string
	CADirURL   string `koanf:"ca_dir_url"`
	CAInsecure bool   `koanf:"ca_insecure"`
}

func (c TLSLetsEncryptConfig) Validate() valid.Problems {
	return valid.Validate(
		valid.String("email", c.Email, valid.Required),
		valid.String("directory", c.Directory, valid.Required, directory),
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

func (c GeoIPConfig) Validate() valid.Problems {
	if !c.Enabled {
		return nil
	}
	return valid.Validate(
		valid.String("city", c.City, valid.Required, file),
		valid.String("asn", c.ASN, valid.Required, file),
	)
}
