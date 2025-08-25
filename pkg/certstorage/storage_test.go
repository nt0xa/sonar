package certstorage_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/v3/acme"
	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/registration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/pkg/certstorage"
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

	_, err := certstorage.New("not-exist")
	assert.Error(t, err)
}

func Test_New_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := certstorage.New("test", certstorage.FilePerm(0600))
	assert.NoError(t, err)
}

func Test_SaveLoadAccount(t *testing.T) {
	setup(t)
	defer teardown(t)

	s, err := certstorage.New("test")
	require.NoError(t, err)

	key, _ := certcrypto.GeneratePrivateKey(certcrypto.EC256)

	accSave := certstorage.Account{
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

	// Remove account file
	err = os.Remove("test/account.json")
	require.NoError(t, err)

	_, err = s.LoadAccount()
	assert.Error(t, err)

	// Remove root directory
	err = os.RemoveAll("test")
	require.NoError(t, err)

	err = s.SaveAccount(&accSave)
	assert.Error(t, err)
}

func Test_SaveLoadCert(t *testing.T) {
	setup(t)
	defer teardown(t)

	s, err := certstorage.New("test")
	require.NoError(t, err)

	cert, err := genCert()
	require.NoError(t, err)

	err = s.SaveCertificate(cert)
	require.NoError(t, err)

	certLoad, err := s.LoadCertificate()
	require.NoError(t, err)

	assert.NotNil(t, certLoad.Leaf)

	assert.Equal(t, cert.Certificate[0], certLoad.Certificate[0])
	assert.Equal(t, cert.PrivateKey, certLoad.PrivateKey)

	// Remove key file
	err = os.Remove("test/tls.key")
	require.NoError(t, err)

	_, err = s.LoadCertificate()
	assert.Error(t, err)

	// Remove root directory
	err = os.RemoveAll("test")
	require.NoError(t, err)

	err = s.SaveCertificate(cert)
	assert.Error(t, err)
}

func genCert() (*tls.Certificate, error) {
	rootKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	var rootTemplate = x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country:      []string{"SE"},
			Organization: []string{"Company Co."},
			CommonName:   "Root CA",
		},
		NotBefore:             time.Now().Add(-10 * time.Second),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	rootDerBytes, err := x509.CreateCertificate(rand.Reader, &rootTemplate, &rootTemplate, &rootKey.PublicKey, rootKey)
	if err != nil {
		return nil, err
	}

	rootPemBytes := &bytes.Buffer{}
	err = pem.Encode(rootPemBytes, &pem.Block{Type: "CERTIFICATE", Bytes: rootDerBytes})
	if err != nil {
		return nil, err
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

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

	certDerBytes, err := x509.CreateCertificate(rand.Reader, &template, &rootTemplate, &key.PublicKey, rootKey)
	if err != nil {
		return nil, err
	}

	certPemBytes := &bytes.Buffer{}
	err = pem.Encode(certPemBytes, &pem.Block{Type: "CERTIFICATE", Bytes: certDerBytes})
	if err != nil {
		return nil, err
	}

	keyPemBytes := &bytes.Buffer{}
	err = pem.Encode(keyPemBytes, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err != nil {
		return nil, err
	}

	full := append(certPemBytes.Bytes()[:], rootPemBytes.Bytes()[:]...)

	cert, err := tls.X509KeyPair(full, keyPemBytes.Bytes())
	if err != nil {
		return nil, err
	}

	return &cert, nil
}
