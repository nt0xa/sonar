package tls

import (
	"crypto/tls"
	"fmt"
	"sync"

	"github.com/go-acme/lego/v3/challenge"

	"github.com/bi-zone/sonar/internal/tls/certmgr"
	"github.com/bi-zone/sonar/internal/utils/logger"
)

type TLS struct {
	cfg *Config
	cm  *certmgr.CertMgr
	wg  sync.WaitGroup
}

func New(cfg *Config, log logger.StdLogger, domain string, provider challenge.Provider) (*TLS, error) {
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
