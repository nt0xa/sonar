package dns_test

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bi-zone/sonar/pkg/server/dns"
	"github.com/golang/mock/gomock"
)

var (
	srv *dns.Server
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
	wg.Add(1)

	srv = dns.New(":1053", "sonar.local", net.IPv4(127, 0, 0, 1),
		dns.NotifyStartedFunc(func() {
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

type containsMatcher struct{ s string }

func (m containsMatcher) Matches(value interface{}) bool {
	s := ""
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return false
	}

	return strings.Contains(s, m.s)
}

func (m containsMatcher) String() string {
	return fmt.Sprintf("contains %q", m.s)
}

func Contains(s string) gomock.Matcher { return containsMatcher{s} }
