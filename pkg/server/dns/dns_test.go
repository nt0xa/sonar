package dns_test

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/mock"

	dnssrv "github.com/bi-zone/sonar/pkg/server/dns"
)

var (
	srv      *dnssrv.Server
	handler  *HandlerMock
	notifier *NotifierMock
)

type HandlerMock struct {
	mock.Mock
}

func (m *HandlerMock) HandleFunc(w dns.ResponseWriter, r *dns.Msg) {
	m.Called(w, r)
	msg := &dns.Msg{}
	msg.SetReply(r)
	w.WriteMsg(msg)
}

type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) Notify(remoteAddr net.Addr, data []byte, meta map[string]interface{}) {
	m.Called(remoteAddr, data, meta)
}

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
	wg.Add(1)

	handler = &HandlerMock{}
	notifier = &NotifierMock{}

	srv = dnssrv.New(":1053", handler.HandleFunc,
		dnssrv.NotifyRequestFunc(notifier.Notify),
		dnssrv.NotifyStartedFunc(func() {
			wg.Done()
		}))

	go func() {
		if err := srv.ListenAndServe(); err != nil {
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
