package httpdb_test

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-testfixtures/testfixtures"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/testutils"
	"github.com/bi-zone/sonar/internal/testutils/httpt"
	"github.com/bi-zone/sonar/pkg/httpx"
)

// Flags
var (
	proxy string
)

var _ = func() bool {
	testing.Init()
	return true
}()

func init() {
	flag.StringVar(&proxy, "test.proxy", "", "Enables verbose HTTP proxy.")
	flag.Parse()
}

func notify(remoteAddr net.Addr, data []byte, meta map[string]interface{}) {}

var (
	db      *database.DB
	tf      *testfixtures.Context
	srvHTTP httpx.Server

	tlsConfig *tls.Config
	srvHTTPS  httpx.Server

	g = testutils.Globals(
		testutils.DB(&database.Config{
			DSN:        os.Getenv("SONAR_DB_DSN"),
			Migrations: "../../internal/database/migrations",
		}, &db),
		testutils.Fixtures(&db, "../../internal/database/fixtures", &tf),
		testutils.TLSConfig("../../test/cert.pem", "../../test/key.pem", &tlsConfig),
		testutils.HTTPX(&db, notify, nil, &srvHTTP),
		testutils.HTTPX(&db, notify, &tlsConfig, &srvHTTPS),
	)
)

func TestMain(m *testing.M) {
	testutils.TestMain(m, g)
}

func setup(t *testing.T) {
	err := tf.Load()
	require.NoError(t, err)
}

func teardown(t *testing.T) {}

func TestHTTPDB(t *testing.T) {
	var tests = []struct {
		host     string
		method   string
		path     string
		query    map[string]string
		headers  map[string]string
		typ      string
		body     map[string]httpt.FormField
		matchers []httpt.ResponseMatcher
	}{

		{
			"c1da9f3d.sonar.local",
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
			[]httpt.ResponseMatcher{
				httpt.Header("test", httpt.Equal("test")),
				httpt.Header("dynamic", httpt.Equal("Header: test-header")),
				httpt.Body(httpt.Contains("Body: query-param")),
			},
		},
		{
			"c1da9f3d.sonar.local",
			"POST",
			"/post",
			map[string]string{
				"test": "query-param",
			},
			map[string]string{
				"Test": "test-header",
			},
			"form",
			map[string]httpt.FormField{
				"test": httpt.StringField("test-body"),
			},
			[]httpt.ResponseMatcher{
				httpt.Code(201),
				httpt.Body(httpt.Contains("Body: test-body")),
			},
		},
		{
			"c1da9f3d.sonar.local",
			"DELETE",
			"/delete",
			map[string]string{
				"test": "query-param",
			},
			nil,
			"",
			nil,
			[]httpt.ResponseMatcher{
				httpt.Code(200),
				httpt.Body(httpt.Contains("DELETE")),
				httpt.Body(httpt.Contains("/delete")),
				httpt.Body(httpt.Contains("test=query-param")),
				httpt.Body(httpt.Contains("/delete?test=query-param")),
			},
		},
		{
			"c1da9f3d.sonar.local",
			"PUT",
			"/route/route-param",
			nil,
			nil,
			"",
			nil,
			[]httpt.ResponseMatcher{
				httpt.Code(200),
				httpt.Body(httpt.Contains("Route: route-param")),
			},
		},
	}

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
				setup(t)
				defer teardown(t)

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
							fw, err := value.Writer(w, name)
							require.NoError(t, err)

							_, err = io.Copy(fw, value.Reader())
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

				// Set host
				if tt.host != "" {
					req.Host = tt.host
				}

				// Set query parameters.
				q := req.URL.Query()
				for name, value := range tt.query {
					q.Add(name, value)
				}
				req.URL.RawQuery = q.Encode()

				// Client parameters.
				tr := &http.Transport{
					DisableKeepAlives: true,
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
				defer tr.CloseIdleConnections()

				if proxy != "" {
					url, err := url.Parse(proxy)
					require.NoError(t, err)
					tr.Proxy = http.ProxyURL(url)
				}

				client := &http.Client{
					Timeout:   5 * time.Second,
					Transport: tr,
				}

				//
				// Send request
				//

				res, err := client.Do(req)
				require.NoError(t, err)

				for _, m := range tt.matchers {
					m(t, res)
				}

				defer res.Body.Close()
			})
		}
	}
}
