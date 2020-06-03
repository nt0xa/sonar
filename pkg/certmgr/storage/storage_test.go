package storage_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/v3/acme"
	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/registration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/pkg/certmgr/storage"
)

func setup(t *testing.T) {
	err := os.Mkdir("test", 0700)
	require.NoError(t, err)
}

func teardown(t *testing.T) {
	err := os.RemoveAll("test")
	require.NoError(t, err)
}

func Test_New_Fail(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := storage.New("not-exist")
	assert.Error(t, err)
}

func Test_New_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := storage.New("test", storage.FilePerm(0600))
	assert.NoError(t, err)
}

func Test_SaveLoadAccount(t *testing.T) {
	setup(t)
	defer teardown(t)

	s, err := storage.New("test")
	require.NoError(t, err)

	key, _ := certcrypto.GeneratePrivateKey(certcrypto.EC256)

	accSave := storage.Account{
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

	err = s.SaveAccount(&accSave)
	require.NoError(t, err)

	accLoad, err := s.LoadAccount()
	require.NoError(t, err)

	assert.Equal(t, accSave, *accLoad)
}

func Test_SaveLoadCert(t *testing.T) {
	setup(t)
	defer teardown(t)

	s, err := storage.New("test")
	require.NoError(t, err)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDerBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, err)

	certPemBytes := &bytes.Buffer{}
	err = pem.Encode(certPemBytes, &pem.Block{Type: "CERTIFICATE", Bytes: certDerBytes})
	require.NoError(t, err)

	keyPemBytes := &bytes.Buffer{}
	err = pem.Encode(keyPemBytes, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	require.NoError(t, err)

	cert, err := tls.X509KeyPair(certPemBytes.Bytes(), keyPemBytes.Bytes())
	require.NoError(t, err)

	err = s.SaveCertificate(&cert)
	require.NoError(t, err)

	certLoad, err := s.LoadCertificate()
	require.NoError(t, err)

	assert.Equal(t, certDerBytes, certLoad.Certificate[0])
	assert.Equal(t, key, certLoad.PrivateKey)
}
