package certmgr

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"github.com/go-acme/lego/v3/certificate"
	"github.com/go-acme/lego/v3/challenge"
	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/log"
	"github.com/go-acme/lego/v3/registration"

	"github.com/nt0xa/sonar/pkg/certstorage"
)

type CertMgr struct {
	email   string
	domains []string

	options *options

	storage  *certstorage.Storage
	provider challenge.Provider

	cert *tls.Certificate

	mu   sync.RWMutex
	once sync.Once

	log StdLogger
}

func New(root string, email string, domains []string, provider challenge.Provider,
	opts ...Option) (*CertMgr, error) {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	storage, err := certstorage.New(root)

	if err != nil {
		return nil, fmt.Errorf("fail to create storage: %w", err)
	}

	cm := &CertMgr{
		email:    email,
		domains:  domains,
		options:  &options,
		storage:  storage,
		provider: provider,
		log:      options.log,
	}

	// Set lego logger
	log.Logger = options.log

	return cm, nil
}

func (cm *CertMgr) SetOption(opt Option) {
	opt(cm.options)
}

func (cm *CertMgr) GetTLSConfig() *tls.Config {
	return &tls.Config{
		GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			return cm.getCert(), nil
		},
	}
}

func (cm *CertMgr) Start() error {
	// Try to load certificate from storage
	if cert, _ := cm.storage.LoadCertificate(); cert != nil {
		cm.setCert(cert)
	}

	// Load account from storage or create a new one
	acc, err := cm.loadOrCreateAccount()
	if err != nil {
		return fmt.Errorf("fail to load/create account: %w", err)
	}

	ticker := time.NewTicker(cm.options.renewInterval)

	for ; true; <-ticker.C {
		now := cm.options.timeNow()

		if cert := cm.getCert(); cert != nil &&
			cert.Leaf.NotAfter.After(now) &&
			cert.Leaf.NotAfter.Sub(now) > cm.options.renewThreshold {
			cm.notifyReady()
			continue
		}

		cert, err := cm.obtainCertificate(acc)
		if err != nil {
			return fmt.Errorf("fail to obtain certificate: %w", err)
		}

		// Store certificate in storage to use in future
		if err := cm.storage.SaveCertificate(cert); err != nil {
			return fmt.Errorf("fail to save certificate: %w", err)
		}

		cm.setCert(cert)
		cm.notifyReady()
	}

	return nil
}

func (cm *CertMgr) obtainCertificate(acc registration.User) (*tls.Certificate, error) {

	client, err := newClient(acc, cm.options.keyType,
		cm.options.caDirURL, cm.options.caInsecure)
	if err != nil {
		return nil, fmt.Errorf("fail to create lego client for account: %w", err)
	}

	request := certificate.ObtainRequest{
		Domains: cm.domains,
		Bundle:  true,
	}

	if err := client.Challenge.SetDNS01Provider(cm.provider,
		dns01.DisableCompletePropagationRequirement()); err != nil {
		return nil, err
	}

	res, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, err
	}

	full := append(res.Certificate[:], res.IssuerCertificate...)

	cert, err := tls.X509KeyPair(full, res.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("fail to create tls certificate: %w", err)
	}

	// tls.X509KeyPair parses certificate but doesn't store it anywhere
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("fail to parse certificate: %w", err)
	}

	cert.Leaf = x509Cert

	return &cert, nil
}

func (cm *CertMgr) loadOrCreateAccount() (*certstorage.Account, error) {
	var (
		acc *certstorage.Account
		err error
	)

	// Try to load account from storage
	if acc, _ = cm.storage.LoadAccount(); acc == nil {

		// There is no account in storage — create a new one
		acc, err = newAccount(cm.email, cm.options.keyType)
		if err != nil {
			return nil, fmt.Errorf("fail to create new account: %w", err)
		}
	}

	if acc.Registration == nil {

		// Threre is no registration info in account, we need to register
		client, err := newClient(acc, cm.options.keyType,
			cm.options.caDirURL, cm.options.caInsecure)
		if err != nil {
			return nil, fmt.Errorf("fail to create lego client for account: %w", err)
		}

		reg, err := registerAccount(client, acc)
		if err != nil {
			return nil, fmt.Errorf("fail to register account: %w", err)
		}

		acc.Registration = reg
	}

	// Store account in storage to use in future
	if err := cm.storage.SaveAccount(acc); err != nil {
		return acc, fmt.Errorf("fail to save account: %w", err)
	}

	return acc, nil
}

func (cm *CertMgr) notifyReady() {
	cm.once.Do(func() {
		if cm.options.notifyReadyFunc != nil {
			cm.options.notifyReadyFunc()
		}
	})
}

func (cm *CertMgr) setCert(cert *tls.Certificate) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.cert = cert
}

func (cm *CertMgr) getCert() *tls.Certificate {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.cert
}
