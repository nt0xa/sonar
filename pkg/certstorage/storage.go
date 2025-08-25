package certstorage

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-acme/lego/v3/certcrypto"
)

const (
	accountDataFileName = "account.json"
	accountKeyFileName  = "account.key"
	keyFileName         = "tls.key"
	certFileName        = "tls.crt"
	caFileName          = "ca-%d.crt"
)

type Storage struct {
	rootPath string
	options  *options
}

func New(root string, opts ...Option) (*Storage, error) {
	options := defaultOptions

	for _, fopt := range opts {
		fopt(&options)
	}

	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, fmt.Errorf("root path %q doesn't exist: %w", root, err)
	}

	return &Storage{
		rootPath: root,
		options:  &options,
	}, nil
}

func (s *Storage) SaveAccount(account *Account) error {
	jsonBytes, err := json.MarshalIndent(account, "", "  ")
	if err != nil {
		return fmt.Errorf("fail to encode account data: %w", err)
	}

	if err := os.WriteFile(
		filepath.Join(s.rootPath, accountDataFileName),
		jsonBytes,
		s.options.filePerm,
	); err != nil {
		return fmt.Errorf("fail to save account data: %w", err)
	}

	pemBytes, err := pemEncode(account.Key)
	if err != nil {
		return fmt.Errorf("fail to pem encode account key: %w", err)
	}

	if err := os.WriteFile(
		filepath.Join(s.rootPath, accountKeyFileName),
		pemBytes,
		s.options.filePerm,
	); err != nil {
		return fmt.Errorf("fail to save account key: %w", err)
	}

	return nil
}

func (s *Storage) LoadAccount() (*Account, error) {
	jsonBytes, err := os.ReadFile(filepath.Join(s.rootPath, accountDataFileName))

	if err != nil {
		return nil, fmt.Errorf("fail to load account data: %w", err)
	}

	var account Account

	if err := json.Unmarshal(jsonBytes, &account); err != nil {
		return nil, fmt.Errorf("fail to decode account data: %w", err)
	}

	keyBytes, err := os.ReadFile(filepath.Join(s.rootPath, accountKeyFileName))
	if err != nil {
		return nil, fmt.Errorf("fail to read account key: %w", err)
	}

	keyBlock, _ := pem.Decode(keyBytes)

	var key crypto.PrivateKey

	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		key, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		key, err = x509.ParseECPrivateKey(keyBlock.Bytes)
	}

	if err != nil {
		return nil, fmt.Errorf("fail to parse account key: %w", err)
	}

	account.Key = key

	return &account, nil
}

func (s *Storage) SaveCertificate(cert *tls.Certificate) error {
	certPemBytes, err := pemEncode(certcrypto.DERCertificateBytes(cert.Certificate[0]))
	if err != nil {
		return fmt.Errorf("fail to pem encode certificate: %w", err)
	}

	keyPemBytes, err := pemEncode(cert.PrivateKey)
	if err != nil {
		return fmt.Errorf("fail to pem encode private key: %w", err)
	}

	if err := os.WriteFile(
		filepath.Join(s.rootPath, certFileName),
		certPemBytes,
		s.options.filePerm,
	); err != nil {
		return fmt.Errorf("fail to save certificate: %w", err)
	}

	if err := os.WriteFile(
		filepath.Join(s.rootPath, keyFileName),
		keyPemBytes,
		s.options.filePerm,
	); err != nil {
		return fmt.Errorf("fail to save key: %w", err)
	}

	for i := 1; i < len(cert.Certificate); i++ {
		pemBytes, err := pemEncode(certcrypto.DERCertificateBytes(cert.Certificate[i]))
		if err != nil {
			return fmt.Errorf("fail to pem encode ca certificate: %w", err)
		}

		if err := os.WriteFile(
			filepath.Join(s.rootPath, fmt.Sprintf(caFileName, i)),
			pemBytes,
			s.options.filePerm,
		); err != nil {
			return fmt.Errorf("fail to save ca certificate: %w", err)
		}
	}

	return nil
}

func (s *Storage) LoadCertificate() (*tls.Certificate, error) {

	certPemBytes, err := os.ReadFile(filepath.Join(s.rootPath, certFileName))
	if err != nil {
		return nil, fmt.Errorf("fail to load certificate: %w", err)
	}

	keyPemBytes, err := os.ReadFile(filepath.Join(s.rootPath, keyFileName))
	if err != nil {
		return nil, fmt.Errorf("fail to load key: %w", err)
	}

	i := 1

	for {
		path := filepath.Join(s.rootPath, fmt.Sprintf(caFileName, i))

		if !fileExists(path) {
			break
		}

		pemBytes, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("fail to load certificate: %w", err)
		}

		certPemBytes = append(certPemBytes, pemBytes...)

		i += 1
	}

	cert, err := tls.X509KeyPair(certPemBytes, keyPemBytes)
	if err != nil {
		return nil, fmt.Errorf("fail to create x509 keypair: %w", err)
	}

	// tls.X509KeyPair parses certificate but doesn't store it anywhere
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("fail to parse certificate: %w", err)
	}

	cert.Leaf = x509Cert

	return &cert, nil
}
