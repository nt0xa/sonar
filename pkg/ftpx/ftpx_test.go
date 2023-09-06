package ftpx_test

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/russtone/sonar/pkg/ftpx"
	"github.com/russtone/sonar/pkg/netx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

	options := []ftpx.Option{
		ftpx.NotifyStartedFunc(wg.Done),
		ftpx.ListenerWrapper(func(l net.Listener) net.Listener {
			return &netx.TimeoutListener{
				Listener: &netx.MaxBytesListener{
					Listener: l,
					MaxBytes: 1 << 20,
				},
				IdleTimeout: 5 * time.Second,
			}
		}),
		ftpx.OnClose(func(e *ftpx.Event) {
			notifier.Notify(e.RemoteAddr, e.Log, map[string]interface{}{})
		}),
	}

	go func() {
		srv := ftpx.New("127.0.0.1:10021", options...)

		if err := srv.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("fail to start server: %s", err))
			os.Exit(1)
		}
	}()

	go func() {
		cert, err := tls.LoadX509KeyPair(
			"../../test/cert.pem",
			"../../test/key.pem",
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("fail to read cert and key: %s", err))
			os.Exit(1)
		}

		options := append(options, ftpx.TLSConfig(&tls.Config{
			Certificates: []tls.Certificate{cert},
		}))
		srv := ftpx.New("127.0.0.1:10022", options...)

		if err := srv.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("fail to start server: %s", err))
			os.Exit(1)
		}
	}()

	if WaitTimeout(&wg, 30*time.Second) {
		fmt.Fprintf(os.Stderr, "timeout waiting for server to start")
		os.Exit(1)
	}

	os.Exit(m.Run())
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
				conn, err = tls.Dial("tcp", "127.0.0.1:10022", tlsConfig)
				require.NoError(st, err)
			} else {
				// Simple TCP connection
				conn, err = net.Dial("tcp", "127.0.0.1:10021")
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
					}),
					mock.Anything,
				).
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
