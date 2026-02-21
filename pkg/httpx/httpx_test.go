package httpx_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"

	"github.com/nt0xa/sonar/pkg/httpx"
)

var (
	notifier = &NotifierMock{}
)

type NotifierMock struct {
	mock.Mock
}

func (m *NotifierMock) Notify(remoteAddr net.Addr, data []byte, secure bool) {
	m.Called(remoteAddr.String(), string(data), secure)
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
	wg.Add(3)

	h := http.TimeoutHandler(
		httpx.BodyReaderHandler(
			httpx.MaxBytesHandler(
				httpx.NotifyHandler(
					func(
						ctx context.Context,
						remoteAddr net.Addr,
						receivedAt *time.Time,
						secure bool,
						read, written, combined []byte,
						meta *httpx.Meta,
					) {
						notifier.Notify(remoteAddr, combined, secure)
					},
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "text/html; charset=utf-8")
						w.WriteHeader(200)
						_, _ = w.Write([]byte("<html><body>test</body></html>"))
					}),
				),
				1<<20,
			),
			1<<20,
		),
		5*time.Second,
		"timeout",
	)

	options := []httpx.Option{
		httpx.NotifyStartedFunc(wg.Done),
	}

	go func() {
		srv := httpx.New("127.0.0.1:1080", h, options...)
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

		options := append(options, httpx.TLSConfig(&tls.Config{
			Certificates: []tls.Certificate{cert},
		}))
		srv := httpx.New("127.0.0.1:1443", h, options...)

		if err := srv.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, "fail to start server: %s", err)
			os.Exit(1)
		}
	}()

	// h2c server: HTTP/2 over cleartext on port 1082.
	go func() {
		opts := append(options, httpx.H2C())
		srv := httpx.New("127.0.0.1:1082", h, opts...)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(fmt.Errorf("fail to start h2c server: %w", err))
		}
	}()

	if WaitTimeout(&wg, 30*time.Second) {
		fmt.Fprintf(os.Stderr, "timeout waiting for server to start")
		os.Exit(1)
	}

	os.Exit(m.Run())
}

type stringField struct {
	s string
}

func (f *stringField) String() string {
	return f.s
}

func (f *stringField) Reader() io.Reader {
	return strings.NewReader(f.s)
}

type fileField struct {
	Name string
	f    *stringField
}

func (f *fileField) String() string {
	return f.f.String()
}

func (f *fileField) Reader() io.Reader {
	return f.f.Reader()
}

type Field interface {
	String() string
	Reader() io.Reader
}

func String(s string) Field {
	return &stringField{s}
}

func File(name string, data string) Field {
	return &fileField{name, &stringField{data}}
}

func TestHTTPX(t *testing.T) {
	var tests = []struct {
		method  string
		path    string
		query   map[string]string
		headers map[string]string
		typ     string
		body    map[string]Field
	}{
		{
			"GET",
			"/get",
			map[string]string{
				"test": "query-param",
			},
			map[string]string{
				"Test": "test-header",
			},
			"",
			nil,
		},
		{
			"POST",
			"/post",
			map[string]string{
				"test": "query-param",
			},
			map[string]string{
				"Test": "test-header",
			},
			"form",
			map[string]Field{
				"test": String("test-body"),
			},
		},
		{
			"POST",
			"/post",
			map[string]string{
				"test": "query-param",
			},
			map[string]string{
				"Test": "test-header",
			},
			"multipart",
			map[string]Field{
				"test": String("test-body"),
				"file": File("file", strings.Repeat("C", 1000)),
			},
		},
	}

	var wg sync.WaitGroup

	wg.Add(len(tests) * 2)

	for _, tt := range tests {
		for _, isTLS := range []bool{false, true} {
			var proto string

			if isTLS {
				proto = "HTTPS"
			} else {
				proto = "HTTP"
			}

			name := fmt.Sprintf("%s/%s", proto, tt.method)

			t.Run(name, func(t *testing.T) {

				// Strings that must be present in the notification.
				contains := make([]string, 0)
				contains = append(contains, tt.path)
				for _, value := range tt.query {
					contains = append(contains, value)
				}
				for _, value := range tt.headers {
					contains = append(contains, value)
				}
				for _, value := range tt.body {
					contains = append(contains, value.String())
				}

				//
				// URI
				//

				var uri string

				if isTLS {
					uri = "https://127.0.0.1:1443"
				} else {
					uri = "http://127.0.0.1:1080"
				}

				//
				// Body
				//

				var (
					body        io.Reader
					contentType string
				)

				if tt.method == "POST" {

					switch tt.typ {
					case "form":
						buf := new(bytes.Buffer)
						params := url.Values{}
						for name, value := range tt.body {
							params.Set(name, value.String())
						}
						buf.WriteString(params.Encode())
						body = buf
						contentType = "application/x-www-form-urlencoded"

					case "multipart":
						buf := new(bytes.Buffer)
						w := multipart.NewWriter(buf)
						for name, value := range tt.body {

							var (
								fw  io.Writer
								r   io.Reader
								err error
							)

							switch v := value.(type) {
							case *stringField:
								fw, err = w.CreateFormField(name)
							case *fileField:
								fw, err = w.CreateFormFile(name, v.Name)
							default:
								t.Errorf("invalid type %T for body field %q", value, name)
							}

							require.NoError(t, err)
							r = value.Reader()

							_, err = io.Copy(fw, r)
							require.NoError(t, err)
						}
						body = buf
						contentType = w.FormDataContentType()

					default:
						t.Errorf("invalid body type %q", tt.typ)
					}
				}

				//
				// Build request
				//

				req, err := http.NewRequest(tt.method, uri+tt.path, body)
				require.NoError(t, err)

				// Set headers.
				for name, value := range tt.headers {
					req.Header.Add(name, value)
				}

				if contentType != "" {
					req.Header.Add("Content-Type", contentType)
				}

				// Set query parameters.
				q := req.URL.Query()
				for name, value := range tt.query {
					q.Add(name, value)
				}
				req.URL.RawQuery = q.Encode()

				// Client parameters.
				tr := &http.Transport{
					Dial: func(network, address string) (net.Conn, error) {
						conn, err := net.Dial(network, address)

						if err != nil {
							return nil, err
						}

						// Set up mock calls.
						// It is done here because otherwise it is not possible to
						// know remote address of the connection.
						notifier.
							On("Notify",
								conn.LocalAddr().String(),
								mock.MatchedBy(func(data string) bool {
									for _, s := range contains {
										if !strings.Contains(data, s) {
											return false
										}
									}
									return true
								}),
								isTLS,
							).
							Once().
							Run(func(args mock.Arguments) {
								wg.Done()
							})

						return conn, err
					},
					DisableKeepAlives: true,
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
				defer tr.CloseIdleConnections()

				client := &http.Client{
					Timeout:   5 * time.Second,
					Transport: tr,
				}

				//
				// Send request
				//

				res, err := client.Do(req)
				require.NoError(t, err)

				defer func() {
					_ = res.Body.Close()
				}()

				assert.Equal(t, 200, res.StatusCode)
			})
		}
	}

	if WaitTimeout(&wg, time.Second*5) {
		t.Errorf("timeout")
	}

	notifier.AssertExpectations(t)
}

func TestHTTP2(t *testing.T) {
	var tests = []struct {
		name          string
		url           string
		secure        bool
		contains      []string
		makeTransport func() http.RoundTripper
	}{
		{
			name:     "h2 over TLS",
			url:      "https://127.0.0.1:1443/h2path?q=h2value",
			secure:   true,
			contains: []string{"/h2path", "h2value", "X-Test"},
			makeTransport: func() http.RoundTripper {
				return &http2.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
			},
		},
		{
			name:     "h2c cleartext",
			url:      "http://127.0.0.1:1082/h2cpath?q=h2cvalue",
			secure:   false,
			contains: []string{"/h2cpath", "h2cvalue", "X-Test"},
			makeTransport: func() http.RoundTripper {
				return &http2.Transport{
					AllowHTTP: true,
					DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
						return net.Dial(network, addr)
					},
				}
			},
		},
	}

	var wg sync.WaitGroup
	wg.Add(len(tests))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := tt.makeTransport()

			notifier.
				On("Notify",
					mock.Anything,
					mock.MatchedBy(func(data string) bool {
						for _, s := range tt.contains {
							if !strings.Contains(data, s) {
								return false
							}
						}
						return true
					}),
					tt.secure,
				).
				Once().
				Run(func(args mock.Arguments) {
					wg.Done()
				})

			req, err := http.NewRequest(http.MethodGet, tt.url, nil)
			require.NoError(t, err)
			req.Header.Set("X-Test", "http2-header")

			client := &http.Client{
				Timeout:   5 * time.Second,
				Transport: tr,
			}

			res, err := client.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, 200, res.StatusCode)
			assert.Equal(t, 2, res.ProtoMajor, "expected HTTP/2 response")
		})
	}

	if WaitTimeout(&wg, 5*time.Second) {
		t.Errorf("timeout waiting for HTTP/2 notifications")
	}

	notifier.AssertExpectations(t)
}

func TestKeepAlive(t *testing.T) {
	clientTrace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			assert.False(t, info.Reused, "Connection must not be reused")
		},
	}
	traceCtx := httptrace.WithClientTrace(context.Background(), clientTrace)

	notifier.
		On("Notify",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).
		Times(2)

		// 1st request
	req, err := http.NewRequestWithContext(traceCtx, http.MethodGet, "http://127.0.0.1:1080", nil)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	_, err = io.Copy(io.Discard, res.Body)
	require.NoError(t, err)

	err = res.Body.Close()
	require.NoError(t, err)

	// 2nd request
	req, err = http.NewRequestWithContext(traceCtx, http.MethodGet, "http://127.0.0.1:1080", nil)
	require.NoError(t, err)

	_, err = http.DefaultClient.Do(req)
	require.NoError(t, err)

	http.DefaultClient.CloseIdleConnections()
}
