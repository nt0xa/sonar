package smtp_test

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/pkg/server/mock_server"
	smtpsrv "github.com/bi-zone/sonar/pkg/server/smtp"
)

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

func TestSMTP(t *testing.T) {
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
			ctrl := gomock.NewController(st)
			defer ctrl.Finish()

			m := mock_server.NewMockRequestNotifier(ctrl)
			srv.SetOption(smtpsrv.NotifyRequestFunc(m.Notify))
			srvTLS.SetOption(smtpsrv.NotifyRequestFunc(m.Notify))

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

			// Setup here because now we know local address
			m.
				EXPECT().
				Notify(
					gomock.Eq(conn.LocalAddr()),
					Contains(tt.from, tt.to, tt.subj, tt.body),
					gomock.Eq(map[string]interface{}{}),
				).
				Times(1)

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
		})
	}
}
