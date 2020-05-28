package certmgr

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/lego"
	"github.com/go-acme/lego/v3/registration"

	"github.com/bi-zone/sonar/pkg/certmgr/storage"
)

func newAccount(email string, keyType certcrypto.KeyType) (*storage.Account, error) {

	key, err := certcrypto.GeneratePrivateKey(keyType)
	if err != nil {
		return nil, fmt.Errorf("fail to generate account's private key: %w", err)
	}

	acc := storage.Account{
		Email: email,
		Key:   key,
	}
	return &acc, nil
}

func registerAccount(client *lego.Client, acc registration.User) (*registration.Resource, error) {

	reg, err := client.Registration.Register(registration.RegisterOptions{
		TermsOfServiceAgreed: true,
	})
	if err != nil {
		return nil, fmt.Errorf("fail to register account: %w", err)
	}

	return reg, nil
}

func newClient(acc registration.User, keyType certcrypto.KeyType,
	caDirURL string, caInsecure bool) (*lego.Client, error) {

	config := lego.NewConfig(acc)
	config.CADirURL = caDirURL

	if caInsecure {
		config.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	}

	config.Certificate = lego.CertificateConfig{
		KeyType: keyType,
	}

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("fail to create lego client: %w", err)
	}

	return client, nil
}
