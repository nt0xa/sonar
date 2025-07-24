package server

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
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

	// Load config from environment variables.
	if err := k.Load(env.Provider(".", env.Opt{
		Prefix: "SONAR_",
		TransformFunc: func(k, v string) (string, any) {
			k = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, "SONAR_")), "_", ".")
			return k, v
		},
		EnvironFunc: environFunc,
	}), nil); err != nil {
		return nil, fmt.Errorf("load config from env: %w", err)
	}

	var cfg Config

	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

var ConfigDefaults = map[string]any{
	"tls.letsencrypt.directory":  "./tls",
	"tls.letsencrypt.ca_dir_url": "https://acme-v02.api.letsencrypt.org/directory",
}

type Config struct {
	Domain string
	IP     string

	DB        DBConfig
	DNS       DNSConfig
	TLS       TLSConfig
	Telemetry TelemetryConfig
	Modules   ModulesConfig
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

// Custom

type TLSCustomConfig struct {
	Key  string
	Cert string
}

// LetsEncrypt

type TLSLetsEncryptConfig struct {
	Email      string
	Directory  string
	CADirURL   string `koanf:"ca_dir_url"`
	CAInsecure bool   `koanf:"ca_insecure"`
}

type ModulesConfig struct {
	Enabled []string
}
