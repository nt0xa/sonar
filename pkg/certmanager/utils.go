package certmanager

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"time"

	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/lego"
	"github.com/go-acme/lego/v3/registration"

	"github.com/bi-zone/sonar/pkg/certstorage"
)

func needRenewal(x509Cert *x509.Certificate, domain string, days int) bool {
	if x509Cert.IsCA {
		return false
	}

	if days >= 0 {
		notAfter := int(time.Until(x509Cert.NotAfter).Hours() / 24.0)
		if notAfter > days {
			return false
		}
	}

	return true
}

func newAccount(accountsStorage *certstorage.AccountsStorage, keyType certcrypto.KeyType) (*certstorage.Account, error) {
	privateKey, err := accountsStorage.GetPrivateKey(keyType)
	if err != nil {
		return nil, err
	}

	var account *certstorage.Account
	if accountsStorage.ExistsAccountFilePath() {
		account, err = accountsStorage.LoadAccount(privateKey)
		if err != nil {
			return nil, err
		}
	} else {
		account = &certstorage.Account{Email: accountsStorage.GetUserID(), Key: privateKey}
	}

	return account, nil
}

func newClient(acc registration.User, keyType certcrypto.KeyType, caDirURL string) (*lego.Client, error) {
	config := lego.NewConfig(acc)
	config.CADirURL = caDirURL

	if insecure := os.Getenv(leInsecureSkipVerifyEnvVar); insecure == "true" {
		config.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}

	config.Certificate = lego.CertificateConfig{
		KeyType: keyType,
	}

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
