package certmanager

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/go-acme/lego/v3/certificate"
	"github.com/go-acme/lego/v3/challenge"
	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/lego"
	"github.com/go-acme/lego/v3/registration"

	"github.com/bi-zone/sonar/pkg/certstorage"
)

const (
	leEnvNameEnvVar            = "LE_ENV"
	leCADirUrlEnvVar           = "LE_CA_DIR_URL"
	leInsecureSkipVerifyEnvVar = "LE_INSECURE_SKIP_VERIFY"

	leEnvProduction = "production"
	leEnvStaging    = "staging"
	leEnvTest       = "test"
)

type CertManager struct {
	client          *lego.Client
	options         *options
	caDirURL        string
	accountsStorage *certstorage.AccountsStorage
	certsStorage    *certstorage.CertificatesStorage
}

func New(path, email string, provider challenge.Provider, opts ...Option) (*CertManager, error) {
	// Options
	options := defaultOptions

	for _, fopt := range opts {
		fopt(&options)
	}

	// Set CA dir URL (for debugging)
	caDirURL := lego.LEDirectoryProduction

	switch env := os.Getenv(leEnvNameEnvVar); env {

	case leEnvStaging:
		caDirURL = lego.LEDirectoryStaging

	case leEnvProduction:
		caDirURL = lego.LEDirectoryProduction

	case leEnvTest:
		caDirURL = os.Getenv(leCADirUrlEnvVar)
	}

	accountsStorage, err := certstorage.NewAccountsStorage(path, email, caDirURL)
	if err != nil {
		return nil, err
	}

	certsStorage := certstorage.NewCertificatesStorage(path, true)

	if err := certsStorage.CreateRootFolder(); err != nil {
		return nil, err
	}

	if err := certsStorage.CreateArchiveFolder(); err != nil {
		return nil, err
	}

	account, err := newAccount(accountsStorage, options.keyType)
	if err != nil {
		return nil, err
	}

	client, err := newClient(account, options.keyType, caDirURL)
	if err != nil {
		return nil, err
	}

	if err := client.Challenge.SetDNS01Provider(provider, dns01.DisableCompletePropagationRequirement()); err != nil {
		return nil, err
	}

	if account.Registration == nil {
		reg, err := client.Registration.Register(registration.RegisterOptions{
			TermsOfServiceAgreed: true,
		})
		if err != nil {
			return nil, fmt.Errorf("could not complete registration: %w", err)
		}

		account.Registration = reg

		if err = accountsStorage.Save(account); err != nil {
			return nil, fmt.Errorf("could not save account: %w", err)
		}
	}

	return &CertManager{
		client:          client,
		caDirURL:        caDirURL,
		accountsStorage: accountsStorage,
		certsStorage:    certsStorage,
		options:         &options,
	}, nil
}

func (m *CertManager) Obtain(domains []string) (*tls.Certificate, error) {
	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	res, err := m.client.Certificate.Obtain(request)
	if err != nil {
		return nil, err
	}

	if err := m.certsStorage.SaveResource(res); err != nil {
		return nil, err
	}

	return m.LoadX509KeyPair(domains)
}

func (m *CertManager) Renew(domains []string) (*tls.Certificate, error) {

	if len(domains) == 0 {
		return nil, fmt.Errorf("empty domains")
	}

	domain := domains[0]

	certificates, err := m.certsStorage.ReadCertificate(domain, ".crt")
	if err != nil {
		return nil, fmt.Errorf("error while loading the certificate for domain %q: %w", domain, err)
	}

	cert := certificates[0]

	if !needRenewal(cert, domain, m.options.days) {
		return nil, nil
	}

	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	res, err := m.client.Certificate.Obtain(request)
	if err != nil {
		return nil, err
	}

	if err := m.certsStorage.MoveToArchive(domain); err != nil {
		return nil, err
	}

	if err := m.certsStorage.SaveResource(res); err != nil {
		return nil, err
	}

	return m.LoadX509KeyPair(domains)
}

func (m *CertManager) LoadX509KeyPair(domains []string) (*tls.Certificate, error) {
	if len(domains) == 0 {
		return nil, fmt.Errorf("empty domains")
	}

	domain := domains[0]

	cert, err := m.certsStorage.ReadFile(domain, ".crt")
	if err != nil {
		return nil, fmt.Errorf("error while loading the certificate for domain %q: %w", domain, err)
	}

	key, err := m.certsStorage.ReadFile(domain, ".key")
	if err != nil {
		return nil, fmt.Errorf("error while loading the key for domain %q: %w", domain, err)
	}

	cer, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	return &cer, nil
}
