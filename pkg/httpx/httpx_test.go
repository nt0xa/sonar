package httpx_test

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/testutils"
	"github.com/bi-zone/sonar/pkg/httpx"
)

var (
	db *database.DB

	srvHTTP httpx.Server

	tlsConfig *tls.Config
	srvHTTPS  httpx.Server

	notifier = &testutils.NotifierMock{}

	g = testutils.Globals(
		testutils.TLSConfig("../../test/cert.pem", "../../test/key.pem", &tlsConfig),
		testutils.DB(&database.Config{
			DSN:        os.Getenv("SONAR_DB_DSN"),
			Migrations: "../../internal/database/migrations",
		}, &db),
		testutils.HTTPX(&db, notifier.Notify, nil, &srvHTTP),
		testutils.HTTPX(&db, notifier.Notify, &tlsConfig, &srvHTTPS),
	)
)

func TestMain(m *testing.M) {
	testutils.TestMain(m, g)
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
					uri = "https://localhost:1443"
				} else {
					uri = "http://localhost:1080"
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
								conn.LocalAddr(),
								mock.MatchedBy(func(data []byte) bool {
									for _, s := range contains {
										if !strings.Contains(string(data), s) {
											return false
										}
									}
									return true
								}),
								map[string]interface{}{
									"tls": isTLS,
								}).
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

				defer res.Body.Close()

				assert.Equal(t, 200, res.StatusCode)
			})
		}
	}

	if testutils.WaitTimeout(&wg, time.Second*5) {
		t.Errorf("timeout")
	}

	notifier.AssertExpectations(t)
}
