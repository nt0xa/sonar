package lark

import (
	"github.com/nt0xa/sonar/pkg/valid"
)

type Config struct {
	Admin             string
	AppID             string `koanf:"app_id"`
	AppSecret         string `koanf:"app_secret"`
	VerificationToken string `koanf:"verification_token"`
	Mode              string
	EncryptKey        string `koanf:"encrypt_key"`
	TLSEnabled        bool   `koanf:"tls_enabled"`
	ProxyURL          string `koanf:"proxy_url"`
	ProxyInsecure     bool   `koanf:"proxy_insecure"`
}

const (
	ModeWebhook   = "webhook"
	ModeWebsocket = "websocket"
)

func (c Config) Validate() valid.Problems {
	fields := []valid.Validatable{
		valid.String("admin", c.Admin, valid.Required),
		valid.String("app_id", c.AppID, valid.Required),
		valid.String("app_secret", c.AppSecret, valid.Required),
		valid.String("mode", c.Mode, valid.Required, valid.In(ModeWebhook, ModeWebsocket)),
	}
	if c.Mode == ModeWebhook {
		fields = append(fields,
			valid.String("verification_token", c.VerificationToken, valid.Required),
			valid.String("encrypt_key", c.EncryptKey, valid.Required),
		)
	}
	return valid.Validate(fields...)
}
