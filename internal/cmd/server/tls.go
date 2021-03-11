package server

import (
	"crypto/tls"
	"fmt"
	"sync"

	"github.com/bi-zone/sonar/internal/utils/logger"
	"github.com/bi-zone/sonar/internal/utils/valid"
	"github.com/bi-zone/sonar/pkg/certmgr"
	"github.com/go-acme/lego/v3/challenge"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type TLSConfig struct {
	Type        string               `json:"type"`
	Custom      TLSCustomConfig      `json:"custom"`
	LetsEncrypt TLSLetsEncryptConfig `json:"letsencrypt"`
}

type TLSCustomConfig struct {
	Key  string `json:"key"`
	Cert string `json:"cert"`
}

type TLSLetsEncryptConfig struct {
	Email      string `json:"email"`
	Directory  string `json:"directory"`
	CADirURL   string `json:"caDirUrl" default:"https://acme-v02.api.letsencrypt.org/directory"`
	CAInsecure bool   `json:"caInsecure"`
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

func (c TLSCustomConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Key, validation.Required, validation.By(valid.File)),
		validation.Field(&c.Cert, validation.Required, validation.By(valid.File)),
	)
}

func (c TLSLetsEncryptConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Email, validation.Required),
		validation.Field(&c.Directory, validation.Required, validation.By(valid.Directory)),
	)
}

type TLS struct {
	cfg *TLSConfig
	cm  *certmgr.CertMgr
	wg  sync.WaitGroup
}

func NewTLS(cfg *TLSConfig, log logger.StdLogger, domain string, provider challenge.Provider) (*TLS, error) {
	t := &TLS{
		cfg: cfg,
	}

	if cfg.Type == "letsencrypt" {
		domains := []string{
			domain,                      // domain itseld
			fmt.Sprintf("*.%s", domain), // wildcard
		}

		cm, err := certmgr.New(
			cfg.LetsEncrypt.Directory,
			cfg.LetsEncrypt.Email,
			domains,
			provider,
			certmgr.CADirURL(cfg.LetsEncrypt.CADirURL),
			certmgr.CAInsecure(cfg.LetsEncrypt.CAInsecure),
			certmgr.Logger(log),
			certmgr.NotifyReadyFunc(func() {
				t.wg.Done()
			}),
		)

		if err != nil {
			return nil, fmt.Errorf("fail to init certmgr: %w", err)
		}

		t.cm = cm

		t.wg.Add(1)
	}

	return t, nil
}

func (t *TLS) Start() error {
	if t.cm == nil {
		return nil
	}

	return t.cm.Start()
}

func (t *TLS) Wait() {
	t.wg.Wait()
}

func (t *TLS) GetConfig() (*tls.Config, error) {

	var tlsConfig *tls.Config

	switch t.cfg.Type {

	case "custom":
		cert, err := tls.LoadX509KeyPair(t.cfg.Custom.Cert, t.cfg.Custom.Key)
		if err != nil {
			return nil, fmt.Errorf("fail to load custom certificate and key: %w", err)
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

	case "letsencrypt":
		tlsConfig = t.cm.GetTLSConfig()
	}

	tlsConfig.NextProtos = []string{"http/1.1"}

	return tlsConfig, nil
}
