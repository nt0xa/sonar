package httpdb_test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/russtone/sonar/internal/database"
	"github.com/russtone/sonar/internal/httpdb"
)

var (
	tf  *testfixtures.Loader
	db  *database.DB
	mux http.Handler
)

func TestMain(m *testing.M) {
	var (
		dsn string
		err error
	)

	if dsn = os.Getenv("SONAR_DB_DSN"); dsn == "" {
		fmt.Fprintln(os.Stderr, "empty SONAR_DB_DSN")
		os.Exit(1)
	}

	db, err = database.New(&database.Config{
		DSN:        dsn,
		Migrations: "../database/migrations",
	}, logrus.New())
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to init database: %v\n", err)
		os.Exit(1)
	}

	if err := db.Migrate(); err != nil {
		fmt.Fprintf(os.Stderr, "fail to apply database migrations: %v\n", err)
		os.Exit(1)
	}

	routes := &httpdb.Routes{DB: db, Origin: "sonar.test"}
	mux = httpdb.Handler(routes, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	tf, err = testfixtures.New(
		testfixtures.Database(db.DB.DB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("../database/fixtures"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to load fixtures: %v", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
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
		body     map[string]FormField
		matchers []ResponseMatcher
	}{

		{
			"c1da9f3d.sonar.test",
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
			[]ResponseMatcher{
				Header("test", Equal("test")),
				Header("dynamic", Equal("Header: test-header")),
				Body(Contains("Body: query-param")),
			},
		},
		{
			"c1da9f3d.sonar.test",
			"POST",
			"/post",
			map[string]string{
				"test": "query-param",
			},
			map[string]string{
				"Test": "test-header",
			},
			"form",
			map[string]FormField{
				"test": StringField("test-body"),
			},
			[]ResponseMatcher{
				Code(201),
				Body(Contains("Body: test-body")),
			},
		},
		{
			"c1da9f3d.sonar.test",
			"DELETE",
			"/delete",
			map[string]string{
				"test": "query-param",
			},
			nil,
			"",
			nil,
			[]ResponseMatcher{
				Code(200),
				Body(Contains("DELETE")),
				Body(Contains("/delete")),
				Body(Contains("test=query-param")),
				Body(Contains("/delete?test=query-param")),
			},
		},
		{
			"c1da9f3d.sonar.test",
			"PUT",
			"/route/route-param",
			nil,
			nil,
			"",
			nil,
			[]ResponseMatcher{
				Code(200),
				Body(Contains("Route: route-param")),
			},
		},
	}

	for _, tt := range tests {

		name := fmt.Sprintf("path=%q,method=%q", tt.method, tt.path)

		t.Run(name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			//
			// URI
			//

			var uri string

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

			//
			// Send request
			//

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			for _, m := range tt.matchers {
				m(t, rr)
			}
		})
	}
}

//
// Body fields
//

type stringFormField struct {
	s string
}

func (f *stringFormField) String() string {
	return f.s
}

func (f *stringFormField) Reader() io.Reader {
	return strings.NewReader(f.s)
}

func (f *stringFormField) Writer(w *multipart.Writer, name string) (io.Writer, error) {
	fw, err := w.CreateFormField(name)
	if err != nil {
		return nil, err
	}

	return fw, nil
}

type fileFormField struct {
	Name  string
	inner *stringFormField
}

func (f *fileFormField) String() string {
	return f.inner.String()
}

func (f *fileFormField) Reader() io.Reader {
	return f.inner.Reader()
}

func (f *fileFormField) Writer(w *multipart.Writer, name string) (io.Writer, error) {
	fw, err := w.CreateFormFile(name, f.Name)
	if err != nil {
		return nil, err
	}

	return fw, nil
}

type FormField interface {
	String() string
	Reader() io.Reader
	Writer(*multipart.Writer, string) (io.Writer, error)
}

func StringField(s string) FormField {
	return &stringFormField{s}
}

func FileField(name string, data string) FormField {
	return &fileFormField{name, &stringFormField{data}}
}

//
// Response matchers
//

type Matcher func(*testing.T, interface{})

func Regex(re *regexp.Regexp) Matcher {
	return func(t *testing.T, value interface{}) {
		assert.Regexp(t, re, value)
	}
}

func Equal(expected interface{}) Matcher {
	return func(t *testing.T, value interface{}) {
		assert.EqualValues(t, expected, value)
	}
}

func Contains(s interface{}) Matcher {
	return func(t *testing.T, value interface{}) {
		require.NotNil(t, value)
		assert.Contains(t, value, s)
	}
}

type ResponseMatcher func(*testing.T, *httptest.ResponseRecorder)

func Code(c int) ResponseMatcher {
	return func(t *testing.T, r *httptest.ResponseRecorder) {
		assert.Equal(t, c, r.Code)
	}
}

func Header(key string, match Matcher) ResponseMatcher {
	return func(t *testing.T, r *httptest.ResponseRecorder) {
		header := r.Header().Get(key)
		assert.NotEmpty(t, header)
		match(t, header)
	}
}

func Body(match Matcher) ResponseMatcher {
	return func(t *testing.T, r *httptest.ResponseRecorder) {
		match(t, string(r.Body.String()))
	}
}
