package certmgr_test

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/pkg/certmgr"
	"github.com/bi-zone/sonar/pkg/certmgr/storage"
	"github.com/bi-zone/sonar/pkg/server/dns"
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

	dnsServer *dns.Server
	strg      *storage.Storage
)

func TestMain(m *testing.M) {
	if err := setupGlobals(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	ret := m.Run()

	if err := teardownGlobals(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(ret)
}

func setupGlobals() error {
	dnsStarted := sync.WaitGroup{}
	dnsStarted.Add(1)

	go func() {
		srv := dns.New(":53", testDomains[0], net.ParseIP("127.0.0.1"),
			dns.NotifyStartedFunc(func() {
				dnsStarted.Done()
			}))

		dnsServer = srv

		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed start DNS handler: %s", err.Error())
		}

	}()

	return nil
}

func teardownGlobals() error {
	return nil
}

func setup(t *testing.T) {
	err := os.MkdirAll(testStorageDir, 0700)
	require.NoError(t, err)

	strg, err = storage.New(testStorageDir)
	require.NoError(t, err)
}

func teardown(t *testing.T) {
	err := os.RemoveAll(testStorageDir)
	require.NoError(t, err)
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

	cm, err := certmgr.New(testStorageDir, testEmail, testDomains,
		dnsServer, options...)
	require.NoError(t, err)
	require.NotNil(t, cm)

	go func() {
		err := cm.Start()
		fmt.Println(err)
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

	certStrg, err := strg.LoadCertificate()
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
	time.Sleep(time.Second * 20)

	newCert, err := tlsConf.GetCertificate(nil)
	require.NoError(t, err)
	require.NotNil(t, cert)

	assert.False(t, newCert.Leaf.Equal(cert.Leaf))
	assert.True(t, newCert.Leaf.NotBefore.After(cert.Leaf.NotBefore))

	newCertStrg, err := strg.LoadCertificate()
	require.NoError(t, err)

	// Compare with certificate from storage
	assert.True(t, newCertStrg.Leaf.Equal(newCert.Leaf))
}

func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
