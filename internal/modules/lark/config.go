package lark

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
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
ProxyInsecure     bool `koanf:"proxy_insecure"`
}

const (
	ModeWebhook   = "webhook"
	ModeWebsocket = "websocket"
)

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Admin, validation.Required),
		validation.Field(&c.AppID, validation.Required),
		validation.Field(&c.AppSecret, validation.Required),
		validation.Field(&c.Mode, validation.In(ModeWebhook, ModeWebsocket)),
		validation.Field(&c.VerificationToken, validation.When(c.Mode == ModeWebhook, validation.Required)),
		validation.Field(&c.EncryptKey, validation.When(c.Mode == ModeWebhook, validation.Required)),
	)
}
