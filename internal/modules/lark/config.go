package lark

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Admin             string     `json:"admin"`
	AppID             string     `json:"app_id"`
	AppSecret         string     `json:"app_secret"`
	VerificationToken string     `json:"verification_token"`
	EncryptKey        string     `json:"encrypt_key"`
	TLSEnabled        bool       `json:"tls_enabled" default:"true"`
	ProxyURL          string     `json:"proxy_url"`
	ProxyInsecure     bool       `json:"proxy_insecure"`
	Auth              AuthConfig `json:"auth"`
}

// TODO: think about better auth config
type AuthMode string

const (
	AuthModeAnyone          AuthMode = "anyone"
	AuthModeDepartmentID AuthMode = "department_id"
)

type AuthConfig struct {
	Mode         AuthMode `json:"mode"`
	DepartmentID string   `json:"department_id"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Admin, validation.Required),
		validation.Field(&c.AppID, validation.Required),
		validation.Field(&c.AppSecret, validation.Required),
		validation.Field(&c.VerificationToken, validation.Required),
		validation.Field(&c.EncryptKey),
	)
}
