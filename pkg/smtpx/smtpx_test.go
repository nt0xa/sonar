package smtpx_test

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/bi-zone/sonar/internal/testutils"
	"github.com/bi-zone/sonar/pkg/smtpx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	srv *smtpx.Server

	tlsConfig *tls.Config
	srvTLS    *smtpx.Server

	notifier = &testutils.NotifierMock{}

	g = testutils.Globals(
		testutils.TLSConfig(&tlsConfig),
		testutils.SMTPX(notifier.Notify, &tlsConfig, false, &srv),
		testutils.SMTPX(notifier.Notify, &tlsConfig, true, &srvTLS),
	)
)

func TestMain(m *testing.M) {
	testutils.TestMain(m, g)
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
				On("Notify", conn.LocalAddr(), mock.MatchedBy(func(data []byte) bool {
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
			c, err := smtp.NewClient(conn, "sonar.local")
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

			_, err = fmt.Fprintf(wc, tt.body)
			require.NoError(st, err)

			err = wc.Close()
			require.NoError(st, err)

			c.Quit()

			time.Sleep(time.Microsecond * 500)

			notifier.AssertExpectations(t)
		})
	}
}
