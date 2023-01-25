package ftpx_test

import (
	"crypto/tls"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/russtone/sonar/internal/testutils"
	"github.com/russtone/sonar/pkg/ftpx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	srv *ftpx.Server

	tlsConfig *tls.Config
	srvTLS    *ftpx.Server

	notifier = &testutils.NotifierMock{}

	g = testutils.Globals(
		testutils.TLSConfig(&tlsConfig),
		testutils.FTPX(notifier.Notify, &tlsConfig, false, &srv),
		testutils.FTPX(notifier.Notify, &tlsConfig, true, &srvTLS),
	)
)

func TestMain(m *testing.M) {
	testutils.TestMain(m, g)
}

func TestFTP(t *testing.T) {

	// Commands recorded from XXE in com.sun.org.apache.xerces parser.
	javaXerces := []string{
		"USER username",
		"PASS password",
		"TYPE I",
		"EPSV ALL",
		"EPSV",
		"EPRT |1|172.17.0.5|43337|",
		"RETR filename",
	}

	var tests = []struct {
		commands []string
		user     string
		pass     string
		retr     string
		tls      bool
	}{
		{
			javaXerces,
			"username",
			"password",
			"filename",
			false,
		},
		{
			javaXerces,
			"username",
			"password",
			"filename",
			true,
		},
	}

	for _, tt := range tests {
		var name string

		if tt.tls {
			name = "FTPS"
		} else {
			name = "FTP"
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
				conn, err = tls.Dial("tcp", "localhost:10022", tlsConfig)
				require.NoError(st, err)
			} else {
				// Simple TCP connection
				conn, err = net.Dial("tcp", "localhost:10021")
				require.NoError(st, err)
			}
			defer conn.Close()

			contains := []string{tt.user, tt.pass, tt.retr}

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

			for _, cmd := range tt.commands {
				n, err := conn.Write([]byte(cmd + "\n"))
				require.NoError(t, err)
				require.NotZero(t, n)
			}

			time.Sleep(time.Microsecond * 500)

			notifier.AssertExpectations(t)
		})
	}
}
