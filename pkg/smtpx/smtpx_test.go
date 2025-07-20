package smtpx_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/pkg/netx"
	"github.com/nt0xa/sonar/pkg/smtpx"
)

var (
	notifier = &NotifierMock{}
)

type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) Notify(remoteAddr net.Addr, data []byte, meta map[string]interface{}) {
	m.Called(remoteAddr, data, meta)
}

func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}

func TestMain(m *testing.M) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	options := []smtpx.Option{
		smtpx.NotifyStartedFunc(wg.Done),
		smtpx.ListenerWrapper(func(l net.Listener) net.Listener {
			return &netx.TimeoutListener{
				Listener: &netx.MaxBytesListener{
					Listener: l,
					MaxBytes: 1 << 20,
				},
				IdleTimeout: 5 * time.Second,
			}
		}),
	}

	handler := smtpx.SessionHandler(
		smtpx.Msg{},
		nil,
		func(ctx context.Context, e *smtpx.Event) {
			notifier.Notify(e.RemoteAddr, e.RW, map[string]interface{}{})
		},
	)

	go func() {
		srv := smtpx.New("127.0.0.1:1025", handler, options...)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(fmt.Errorf("fail to start server: %w", err))
		}
	}()

	go func() {
		cert, err := tls.LoadX509KeyPair(
			"../../test/cert.pem",
			"../../test/key.pem",
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fail to read cert and key: %s", err)
			os.Exit(1)
		}

		options := append(options, smtpx.TLSConfig(&tls.Config{
			Certificates: []tls.Certificate{cert},
		}))
		srv := smtpx.New("127.0.0.1:1465", handler, options...)

		if err := srv.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, "fail to start server: %s", err)
			os.Exit(1)
		}
	}()

	if WaitTimeout(&wg, 30*time.Second) {
		fmt.Fprintf(os.Stderr, "timeout waiting for server to start")
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestSMTP(t *testing.T) {
	var tests = []struct {
		from     string
		to       string
		subj     string
		body     string
		startTLS bool
		tls      bool
	}{
		{
			"sender@example.org",
			"recipient@example.net",
			"Test",
			"Test body",
			false,
			false,
		},
		{
			"sender@example.org",
			"recipient@example.net",
			"Test",
			"Test body",
			true,
			false,
		},
		{
			"sender@example.org",
			"recipient@example.net",
			"Test",
			"Test body",
			true,
			true,
		},
	}

	for _, tt := range tests {
		var name string

		if tt.tls {
			name = "SMTPS"
		} else if tt.startTLS {
			name = "SMTP/STARTTLS"
		} else {
			name = "SMTP"
		}

		t.Run(name, func(st *testing.T) {

			// TLS config
			tlsConfig := &tls.Config{
				InsecureSkipVerify: true,
			}

			var (
				conn net.Conn
				err  error
			)

			// Connect to server
			if tt.tls {
				// TLS connection
				conn, err = tls.Dial("tcp", "localhost:1465", tlsConfig)
				require.NoError(st, err)
			} else {
				// Simple TCP connection
				conn, err = net.Dial("tcp", "localhost:1025")
				require.NoError(st, err)
			}

			contains := []string{tt.from, tt.to, tt.subj, tt.body}

			notifier.
				On("Notify",
					mock.MatchedBy(func(addr net.Addr) bool {
						return conn.LocalAddr().String() == addr.String()
					}),
					mock.MatchedBy(func(data []byte) bool {
						for _, s := range contains {
							if !strings.Contains(string(data), s) {
								return false
							}
						}
						return true
					}), mock.Anything).
				Return().
				Once()

			// Create SMTP client
			c, err := smtp.NewClient(conn, "sonar.test")
			require.NoError(st, err)

			// Send "STARTTLS" if required
			if tt.startTLS {
				c.StartTLS(tlsConfig)
			}

			// Set the sender and recipient first
			err = c.Mail(tt.from)
			require.NoError(st, err)

			err = c.Rcpt(tt.to)
			require.NoError(st, err)

			// Send the email body.
			wc, err := c.Data()
			require.NoError(st, err)

			_, err = fmt.Fprintf(wc, "%s", tt.body)
			require.NoError(st, err)

			err = wc.Close()
			require.NoError(st, err)

			c.Quit()

			time.Sleep(time.Microsecond * 500)

			notifier.AssertExpectations(t)
		})
	}
}
