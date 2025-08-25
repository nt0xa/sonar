package certmgr_test

import (
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/pkg/certmgr"
	"github.com/nt0xa/sonar/pkg/certstorage"
)

var (
	testStorageDir = "test"
	testEmail      = "user@example.com"
	testDomains    = []string{"sonar.test"}
	testOptions    = []certmgr.Option{
		certmgr.CAInsecure(true),
		certmgr.RenewInterval(time.Second * 3),
		certmgr.KeyType(certcrypto.EC256),
		certmgr.RenewThreshold(30 * 24 * time.Hour),
	}

	storage *certstorage.Storage
)

func setup(t *testing.T) {
	err := os.MkdirAll(testStorageDir, 0700)
	require.NoError(t, err)

	storage, err = certstorage.New(testStorageDir)
	require.NoError(t, err)
}

func teardown(t *testing.T) {
	err := os.RemoveAll(testStorageDir)
	require.NoError(t, err)
}

type MockProvider struct{}

func (m *MockProvider) Present(domain, token, keyAuth string) error {
	return nil
}

func (m *MockProvider) CleanUp(domain, token, keyAuth string) error {
	return nil
}

func TestCertMgr(t *testing.T) {
	setup(t)
	defer teardown(t)

	wg := sync.WaitGroup{}
	wg.Add(1)

	caDirURL := os.Getenv("CERTMGR_CA_DIR_URL")

	if caDirURL == "" {
		t.Fatal("Empty CERTMGR_CA_DIR_URL")
	}

	options := testOptions
	options = append(options, certmgr.CADirURL(caDirURL))
	options = append(options, certmgr.NotifyReadyFunc(func() {
		wg.Done()
	}))

	cm, err := certmgr.New(testStorageDir, testEmail, testDomains, &MockProvider{}, options...)
	require.NoError(t, err)
	require.NotNil(t, cm)

	go func() {
		err := cm.Start()
		if err != nil {
			wg.Done()
		}
	}()

	wg.Wait()

	tlsConf := cm.GetTLSConfig()
	require.NotNil(t, tlsConf)

	cert, err := tlsConf.GetCertificate(nil)
	require.NoError(t, err)
	require.NotNil(t, cert)

	assert.True(t, cert.Leaf.NotBefore.Before(time.Now()))
	assert.True(t, cert.Leaf.NotAfter.After(time.Now()))

	certStrg, err := storage.LoadCertificate()
	require.NoError(t, err)

	// Compare with certificate from storage
	assert.True(t, certStrg.Leaf.Equal(cert.Leaf))

	// Change time so that cert is expired
	var once sync.Once
	cm.SetOption(certmgr.TestOnlyTimeNow(func() time.Time {
		timeDiff := time.Duration(0)
		once.Do(func() {
			timeDiff = time.Hour * 24 * 365 * 10
		})
		return time.Now().Add(timeDiff)
	}))

	// Wait for certificate to renew
	time.Sleep(time.Second * 5)

	newCert, err := tlsConf.GetCertificate(nil)
	require.NoError(t, err)
	require.NotNil(t, cert)

	assert.False(t, newCert.Leaf.Equal(cert.Leaf))
	assert.True(t, newCert.Leaf.NotBefore.After(cert.Leaf.NotBefore))

	newCertStrg, err := storage.LoadCertificate()
	require.NoError(t, err)

	// Compare with certificate from storage
	assert.True(t, newCertStrg.Leaf.Equal(newCert.Leaf))
}

func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Close()
	}()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
