package smtpx_test

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/bi-zone/sonar/internal/cmd/server"
	"github.com/bi-zone/sonar/internal/testutils"
	"github.com/bi-zone/sonar/pkg/smtpx"
)

var (
	srv    smtpx.Server
	srvTLS smtpx.Server

	notifier = &testutils.NotifierMock{}
)

func TestMain(m *testing.M) {
	if err := setupGlobals(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	ret := m.Run()

	os.Exit(ret)
}

func setupGlobals() error {
	wg := sync.WaitGroup{}
	wg.Add(2)

	cert, err := tls.LoadX509KeyPair("../../test/cert.pem", "../../test/key.pem")
	if err != nil {
		log.Fatal(err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	srv = smtpx.New(
		"localhost:1025",
		server.SMTPListenerWrapper(1<<20, time.Second*5),
		server.SMTPSession("sonar.local", tlsConfig, notifier.Notify),
		smtpx.NotifyStartedFunc(func() {
			wg.Done()
		}),
	)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(fmt.Errorf("fail to start server: %w", err))
		}
	}()

	srvTLS = smtpx.New(
		"localhost:1465",
		server.SMTPListenerWrapper(1<<20, time.Second*5),
		server.SMTPSession("sonar.local", tlsConfig, notifier.Notify),
		smtpx.TLSConfig(tlsConfig),
		smtpx.NotifyStartedFunc(func() {
			wg.Done()
		}),
	)

	go func() {
		if err := srvTLS.ListenAndServe(); err != nil {
			log.Fatal(fmt.Errorf("fail to start server: %w", err))
		}
	}()

	if waitTimeout(&wg, 30*time.Second) {
		return errors.New("timeout waiting for server to start")
	}

	return nil
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
