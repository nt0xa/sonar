package certstorage_test

import (
	"testing"

	"github.com/go-acme/lego/v3/acme"
	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/registration"
	"github.com/stretchr/testify/assert"

	"github.com/bi-zone/sonar/internal/tls/certstorage"
)

func Test_Account(t *testing.T) {
	key, _ := certcrypto.GeneratePrivateKey(certcrypto.EC256)

	acc := certstorage.Account{
		Email: "test@example.com",
		Registration: &registration.Resource{
			Body: acme.Account{
				Status:  "valid",
				Contact: []string{"mailto:test@example.com"},
				Orders:  "https://localhost:14000/list-orderz/3",
			},
			URI: "https://localhost:14000/my-account/3",
		},
		Key: key,
	}

	assert.Equal(t, acc.Email, acc.GetEmail())
	assert.Equal(t, acc.Registration, acc.GetRegistration())
	assert.Equal(t, acc.Key, acc.GetPrivateKey())
}
