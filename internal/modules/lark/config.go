package lark

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Admin             string `mapstructure:"admin"`
	AppID             string `mapstructure:"app_id"`
	AppSecret         string `mapstructure:"app_secret"`
	VerificationToken string `mapstructure:"verification_token"`
	Mode              string `mapstructure:"mode"`
	EncryptKey        string `mapstructure:"encrypt_key"`
	TLSEnabled        bool   `mapstructure:"tls_enabled"`
	ProxyURL          string `mapstructure:"proxy_url"`
	ProxyInsecure     bool   `mapstructure:"proxy_insecure"`
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
