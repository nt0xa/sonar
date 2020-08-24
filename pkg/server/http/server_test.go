package http_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	httpsrv "github.com/bi-zone/sonar/pkg/server/http"
	server_mocks "github.com/bi-zone/sonar/pkg/server/mocks"
)

var tests = []struct {
	method  string
	path    string
	query   map[string]string
	headers map[string]string
	body    map[string]string
	tls     bool
}{
	{
		"GET",
		"/get",
		map[string]string{"test": "query-param"},
		map[string]string{"Test": "test-header"},
		nil,
		false,
	},
	{
		"POST",
		"/post",
		map[string]string{"test": "query-param"},
		map[string]string{"Test": "test-header"},
		map[string]string{"test": "test-body"},
		false,
	},
	{
		"GET",
		"/get",
		map[string]string{"test": "query-param"},
		map[string]string{"Test": "test-header"},
		nil,
		true,
	},
	{
		"POST",
		"/post",
		map[string]string{"test": "query-param"},
		map[string]string{"Test": "test-header"},
		map[string]string{"test": "test-body"},
		true,
	},
}

func TestHTTP(t *testing.T) {

	for _, tt := range tests {
		var proto string

		if tt.tls {
			proto = "HTTPS"
		} else {
			proto = "HTTP"
		}

		name := fmt.Sprintf("%s/%s", proto, tt.method)

		t.Run(name, func(st *testing.T) {
			fmt.Println("START")

			contains := make([]string, 0)
			contains = append(contains, tt.path)
			for _, value := range tt.query {
				contains = append(contains, value)
			}
			for _, value := range tt.headers {
				contains = append(contains, value)
			}
			for _, value := range tt.body {
				contains = append(contains, value)
			}

			ct, cancel := context.WithCancel(context.Background())
			defer cancel()

			notifier := &server_mocks.RequestNotifier{}

			if tt.tls {
				srvTLS.SetOption(httpsrv.NotifyRequestFunc(notifier.Notify))
			} else {
				srv.SetOption(httpsrv.NotifyRequestFunc(notifier.Notify))
			}

			dial := func(ctx context.Context, network, address string) (net.Conn, error) {

				conn, err := (&net.Dialer{}).DialContext(ct, network, address)

				notifier.
					On("Notify", conn.LocalAddr(), mock.MatchedBy(func(data []byte) bool {
						for _, s := range contains {
							if !strings.Contains(string(data), s) {
								return false
							}
						}

						return true
					}), map[string]interface{}{"tls": tt.tls}).
					Once()

				return conn, err
			}

			tr := &http.Transport{
				DialContext:       dial,
				DisableKeepAlives: true,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			defer tr.CloseIdleConnections()

			var uri string

			if tt.tls {
				uri = "https://localhost:1443"
			} else {
				uri = "http://localhost:1080"
			}

			var body io.Reader

			if tt.method == "POST" {
				buf := new(bytes.Buffer)
				params := url.Values{}
				for name, value := range tt.body {
					params.Set(name, value)
				}
				buf.WriteString(params.Encode())
				body = buf
			}

			req, err := http.NewRequest(tt.method, uri+tt.path, body)
			require.NoError(st, err)

			for name, value := range tt.headers {
				req.Header.Add(name, value)
			}

			q := req.URL.Query()
			for name, value := range tt.query {
				q.Add(name, value)
			}
			req.URL.RawQuery = q.Encode()

			client := &http.Client{
				Timeout:   5 * time.Second,
				Transport: tr,
			}
			res, err := client.Do(req)
			require.NoError(st, err)

			defer res.Body.Close()

			assert.Equal(st, 200, res.StatusCode)

			notifier.AssertExpectations(t)
		})
	}
}
