package api_test

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/actionsdb"
	"github.com/russtone/sonar/internal/database"
	"github.com/russtone/sonar/internal/database/models"
	"github.com/russtone/sonar/internal/modules/api"
	"github.com/russtone/sonar/internal/utils/errors"
)

// Flags
var (
	verbose bool
	proxy   string
)

var _ = func() bool {
	testing.Init()
	return true
}()

func init() {
	flag.BoolVar(&verbose, "test.verbose", false, "Enables verbose HTTP printing.")
	flag.StringVar(&proxy, "test.proxy", "", "Enables verbose HTTP proxy.")
	flag.Parse()
}

const (
	AdminToken = "a33bfdbfb3c62feb7ea59314dbd17426"
	User1Token = "50c862e41d059eeca13adc7b276b46b7"
	User2Token = "7001f2d819d3d5fb0b1fd75dd38eb34e"
)

var (
	tf  *testfixtures.Loader
	db  *database.DB
	srv *httptest.Server
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

	log := logrus.New()

	db, err = database.New(&database.Config{
		DSN:        dsn,
		Migrations: "../../database/migrations",
	}, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to init database: %v\n", err)
		os.Exit(1)
	}

	if err := db.Migrate(); err != nil {
		fmt.Fprintf(os.Stderr, "fail to apply database migrations: %v\n", err)
		os.Exit(1)
	}

	acts := actionsdb.New(db, log, "sonar.test")

	tf, err = testfixtures.New(
		testfixtures.Database(db.DB.DB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("../../database/fixtures"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to load fixtures: %v", err)
		os.Exit(1)
	}

	api, err := api.New(&api.Config{Admin: AdminToken}, db, log, nil, acts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to create api server: %v", err)
		os.Exit(1)
	}

	srv = httptest.NewServer(api.Router())

	os.Exit(m.Run())
}

func setup(t *testing.T) {
	err := tf.Load()
	require.NoError(t, err)
}

func teardown(t *testing.T) {}

//
// Matchers
//

type matcher func(*testing.T, interface{})

func regex(re *regexp.Regexp) matcher {
	return func(t *testing.T, value interface{}) {
		assert.Regexp(t, re, value)
	}
}

func withinDuration(d time.Duration) matcher {
	return func(t *testing.T, value interface{}) {
		ts, ok := value.(string)
		assert.True(t, ok, "expected value %+v to be string", value)
		tt, err := time.Parse(time.RFC3339, ts)
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now(), tt, d)
	}
}

func equal(expected interface{}) matcher {
	return func(t *testing.T, value interface{}) {
		assert.EqualValues(t, expected, value)
	}
}

func contains(s interface{}) matcher {
	return func(t *testing.T, value interface{}) {
		require.NotNil(t, value)
		assert.Contains(t, value, s)
	}
}

func notEmpty() matcher {
	return func(t *testing.T, value interface{}) {
		assert.NotEmpty(t, value)
	}
}

func length(l int) matcher {
	return func(t *testing.T, value interface{}) {
		assert.Len(t, value, l)
	}
}

//
// Tests
//

func TestAPI(t *testing.T) {
	tests := []struct {
		method string
		path   string
		query  string
		token  string
		json   string
		schema interface{}
		result map[string]matcher
		status int
	}{

		//
		// Payloads
		//

		// Create

		{
			method: "POST",
			path:   "/payloads",
			token:  User1Token,
			json:   `{"name": "test", "notifyProtocols": ["dns", "smtp"], "storeEvents": true}`,
			schema: actions.PayloadsCreateResult{},
			result: map[string]matcher{
				"$.subdomain": regex(regexp.MustCompile("^[a-f0-9]{8}$")),
				"$.name":      equal("test"),
				"$.notifyProtocols": equal(
					[]interface{}{
						models.ProtoCategoryDNS.String(),
						models.ProtoCategorySMTP.String(),
					},
				),
				"$.storeEvents": equal(true),
				"$.createdAt":   withinDuration(time.Second * 10),
			},
			status: 201,
		},
		{
			method: "POST",
			path:   "/payloads",
			token:  User1Token,
			json:   `{"invalid": 1}`,
			schema: &errors.BaseError{},
			result: map[string]matcher{
				"$.message": contains("format"),
				"$.details": contains("json"),
			},
			status: 400,
		},
		{
			method: "POST",
			path:   "/payloads",
			token:  User1Token,
			json:   `{"name": ""}`,
			schema: &errors.ValidationError{},
			result: map[string]matcher{
				"$.message":     contains("validation"),
				"$.errors.name": notEmpty(),
			},
			status: 400,
		},
		{
			method: "POST",
			path:   "/payloads",
			token:  User1Token,
			json:   `{"name": "payload1"}`,
			schema: &errors.ConflictError{},
			result: map[string]matcher{
				"$.message": contains("conflict"),
			},
			status: 409,
		},

		// List

		{
			method: "GET",
			path:   "/payloads",
			token:  User1Token,
			schema: (actions.PayloadsListResult)(nil),
			result: map[string]matcher{
				"$[0].name": equal("payload1"),
				"$[1].name": equal("payload4"),
			},
			status: 200,
		},
		{
			method: "GET",
			path:   "/payloads",
			query:  "name=payload4",
			token:  User1Token,
			schema: (actions.PayloadsListResult)(nil),
			result: map[string]matcher{
				"$[0].name": equal("payload4"),
			},
			status: 200,
		},

		// Update

		{
			method: "PUT",
			path:   "/payloads/payload1",
			token:  User1Token,
			json:   `{"name":"test", "notifyProtocols": ["smtp"], "storeEvents": false}`,
			schema: actions.PayloadsUpdateResult{},
			result: map[string]matcher{
				"$.name":            equal("test"),
				"$.notifyProtocols": equal([]interface{}{models.ProtoCategorySMTP.String()}),
				"$.storeEvents":     equal(false),
			},
			status: 200,
		},
		{
			method: "PUT",
			path:   "/payloads/payload1",
			token:  User1Token,
			json:   `{"name":"test", "notifyProtocols": ["smtp"], "storeEvents": null}`,
			schema: actions.PayloadsUpdateResult{},
			result: map[string]matcher{
				"$.name":            equal("test"),
				"$.notifyProtocols": equal([]interface{}{models.ProtoCategorySMTP.String()}),
				"$.storeEvents":     equal(true), // Must not be changed
			},
			status: 200,
		},
		{
			method: "PUT",
			path:   "/payloads/payload1",
			token:  User1Token,
			json:   `{"invalid": 1}`,
			schema: &errors.BadFormatError{},
			result: map[string]matcher{
				"$.message": contains("format"),
				"$.details": contains("json"),
			},
			status: 400,
		},
		{
			method: "PUT",
			path:   "/payloads/payload1",
			token:  User1Token,
			json:   `{"name":"test", "notifyProtocols": ["invalid"]}`,
			schema: &errors.ValidationError{},
			result: map[string]matcher{
				"$.message":                contains("validation"),
				"$.errors.notifyProtocols": notEmpty(),
			},
			status: 400,
		},
		{
			method: "PUT",
			path:   "/payloads/invalid",
			token:  User1Token,
			json:   `{"name":"test", "notifyProtocols": ["smtp"]}`,
			schema: &errors.NotFoundError{},
			result: map[string]matcher{
				"$.message": contains("not found"),
			},
			status: 404,
		},

		// Delete

		{
			method: "DELETE",
			path:   "/payloads/payload1",
			token:  User1Token,
			schema: actions.PayloadsDeleteResult{},
			status: 200,
		},
		{
			method: "DELETE",
			path:   "/payloads/invalid",
			token:  User1Token,
			schema: &errors.NotFoundError{},
			result: map[string]matcher{
				"$.message": contains("not found"),
			},
			status: 404,
		},

		// Clear
		{
			method: "DELETE",
			path:   "/payloads",
			token:  User1Token,
			schema: (actions.PayloadsClearResult)(nil),
			result: map[string]matcher{
				"$[0].name": equal("payload1"),
				"$[1].name": equal("payload4"),
			},
			status: 200,
		},

		//
		// DNS records
		//

		// Create

		{
			method: "POST",
			path:   "/dns-records",
			token:  User1Token,
			json:   `{"payloadName": "payload1", "name": "test", "type": "a", "ttl": 100, "values": ["127.0.0.1"], "strategy": "all"}`,
			schema: actions.DNSRecordsCreateResult{},
			result: map[string]matcher{
				"$.name":      equal("test"),
				"$.type":      equal("A"),
				"$.ttl":       equal(100),
				"$.strategy":  equal("all"),
				"$.values":    equal([]interface{}{"127.0.0.1"}),
				"$.createdAt": withinDuration(time.Second * 10),
			},
			status: 201,
		},
		{
			method: "POST",
			path:   "/dns-records",
			token:  User1Token,
			json:   `{"invalid": 1}`,
			schema: &errors.BadFormatError{},
			result: map[string]matcher{
				"$.message": contains("format"),
				"$.details": contains("json"),
			},
			status: 400,
		},
		{
			method: "POST",
			path:   "/dns-records",
			token:  User1Token,
			json:   `{"payloadName": "payload1", "name": ""}`,
			schema: &errors.ValidationError{},
			result: map[string]matcher{
				"$.message": contains("validation"),
			},
			status: 400,
		},
		{
			method: "POST",
			path:   "/dns-records",
			token:  User1Token,
			json:   `{"payloadName": "payload1", "name": "test-a", "type": "a", "ttl": 100, "strategy": "all", "values": ["127.0.0.1"]}`,
			schema: &errors.ValidationError{},
			result: map[string]matcher{
				"$.message": contains("conflict"),
			},
			status: 409,
		},

		// List

		{
			method: "GET",
			path:   "/dns-records/payload1",
			token:  User1Token,
			schema: (actions.DNSRecordsListResult)(nil),
			result: map[string]matcher{
				"$":         length(9),
				"$[0].name": equal("test-a"),
				"$[1].name": equal("test-aaaa"),
			},
			status: 200,
		},
		{
			method: "GET",
			path:   "/dns-records/not-exist",
			token:  User1Token,
			schema: &errors.NotFoundError{},
			result: map[string]matcher{
				"$.message": contains("not found"),
			},
			status: 404,
		},

		// Delete

		{
			method: "DELETE",
			path:   "/dns-records/payload1/1",
			token:  User1Token,
			schema: actions.DNSRecordsDeleteResult{},
			result: map[string]matcher{
				"$.name": equal("test-a"),
			},
			status: 200,
		},
		{
			method: "DELETE",
			path:   "/dns-records/not-exist/1",
			token:  User1Token,
			schema: &errors.NotFoundError{},
			result: map[string]matcher{
				"$.message": contains("not found"),
			},
			status: 404,
		},
		{
			method: "DELETE",
			path:   "/dns-records/payload1/1337",
			token:  User1Token,
			schema: &errors.NotFoundError{},
			result: map[string]matcher{
				"$.message": contains("not found"),
			},
			status: 404,
		},

		// Clear

		{
			method: "DELETE",
			path:   "/dns-records/payload1/",
			token:  User1Token,
			schema: actions.DNSRecordsClearResult{},
			status: 200,
			result: map[string]matcher{
				"$[0].name": equal("test-a"),
				"$[1].name": equal("test-aaaa"),
			},
		},
		{
			method: "DELETE",
			path:   "/dns-records/not-exist/",
			token:  User1Token,
			schema: &errors.NotFoundError{},
			result: map[string]matcher{
				"$.message": contains("not found"),
			},
			status: 404,
		},

		//
		// User
		//

		{
			method: "GET",
			path:   "/profile",
			token:  User1Token,
			schema: actions.ProfileGetResult{},
			result: map[string]matcher{
				"$.name": equal("user1"),
			},
			status: 200,
		},
		{
			method: "GET",
			path:   "/user",
			schema: &errors.BaseError{},
			result: map[string]matcher{
				"$.message": contains("unauthorized"),
			},
			status: 401,
		},
		{
			method: "GET",
			path:   "/user",
			token:  "invalid",
			schema: &errors.BaseError{},
			result: map[string]matcher{
				"$.message": contains("unauthorized"),
			},
			status: 401,
		},

		//
		// Users
		//

		// Create

		{
			method: "POST",
			path:   "/users",
			token:  AdminToken,
			json:   `{"name": "test", "params": {"api.token": "token", "telegram.id": 1234}}`,
			schema: actions.UsersCreateResult{},
			result: map[string]matcher{
				"$.name":                  equal("test"),
				"$.isAdmin":               equal(false),
				`$.params["api.token"]`:   equal("token"),
				`$.params["telegram.id"]`: equal(1234),
				"$.createdAt":             withinDuration(time.Second * 10),
			},
			status: 201,
		},
		{
			method: "POST",
			path:   "/users",
			token:  AdminToken,
			json:   `{"invalid": 1}`,
			schema: &errors.BadFormatError{},
			result: map[string]matcher{
				"$.message": contains("format"),
				"$.details": contains("json"),
			},
			status: 400,
		},
		{
			method: "POST",
			path:   "/users",
			token:  AdminToken,
			json:   `{"name": "user1", "params": {"api.token": "token", "telegram.id": 1234}}`,
			schema: &errors.ConflictError{},
			result: map[string]matcher{
				"$.message": contains("conflict"),
			},
			status: 409,
		},
		{
			method: "POST",
			path:   "/users",
			token:  User1Token,
			schema: &errors.ForbiddenError{},
			result: map[string]matcher{
				"$.message": contains("forbidden"),
			},
			status: 403,
		},

		// Delete

		{
			method: "DELETE",
			path:   "/users/user1",
			token:  AdminToken,
			schema: actions.UsersCreateResult{},
			status: 200,
		},
		{
			method: "DELETE",
			path:   "/users/not-exist",
			token:  AdminToken,
			schema: &errors.NotFoundError{},
			result: map[string]matcher{
				"$.message": contains("not found"),
			},
			status: 404,
		},

		//
		// Events
		//

		// List

		{
			method: "GET",
			path:   "/events/payload1",
			token:  User1Token,
			schema: (actions.EventsListResult)(nil),
			result: map[string]matcher{
				"$":             length(9),
				"$[0].protocol": equal("http"),
				"$[8].protocol": equal("dns"),
			},
			status: 200,
		},

		// Get

		{
			method: "GET",
			path:   "/events/payload1/2",
			token:  User1Token,
			schema: actions.EventsGetResult{},
			result: map[string]matcher{
				"$.protocol": equal("http"),
			},
			status: 200,
		},

		//
		// HTTP routes
		//

		// Create

		{
			method: "POST",
			path:   "/http-routes",
			token:  User1Token,
			json: `{
				"payloadName": "payload1",
				"method": "GET",
				"path": "/test",
				"code": 200,
				"headers": {"Test":["test"]},
				"body": "dGVzdA==",
				"isDynamic": true
			}`,
			schema: actions.HTTPRoutesCreateResult{},
			result: map[string]matcher{
				"$.method":    equal("GET"),
				"$.path":      equal("/test"),
				"$.code":      equal(200),
				"$.headers":   equal(map[string]interface{}{"Test": []interface{}{"test"}}),
				"$.body":      equal("dGVzdA=="),
				"$.isDynamic": equal(true),
				"$.createdAt": withinDuration(time.Second * 10),
			},
			status: 201,
		},
		{
			method: "POST",
			path:   "/http-routes",
			token:  User1Token,
			json:   `{"invalid": 1}`,
			schema: &errors.BadFormatError{},
			result: map[string]matcher{
				"$.message": contains("format"),
				"$.details": contains("json"),
			},
			status: 400,
		},
		{
			method: "POST",
			path:   "/http-routes",
			token:  User1Token,
			json:   `{"payloadName": "payload1", "path": ""}`,
			schema: &errors.ValidationError{},
			result: map[string]matcher{
				"$.message": contains("validation"),
			},
			status: 400,
		},
		{
			method: "POST",
			path:   "/http-routes",
			token:  User1Token,
			json:   `{"payloadName": "payload1", "method": "GET", "path": "/get", "code": 200}`,
			schema: &errors.ValidationError{},
			result: map[string]matcher{
				"$.message": contains("conflict"),
			},
			status: 409,
		},

		// List

		{
			method: "GET",
			path:   "/http-routes/payload1",
			token:  User1Token,
			schema: (actions.HTTPRoutesListResult)(nil),
			result: map[string]matcher{
				"$":         length(5),
				"$[0].path": equal("/get"),
				"$[1].path": equal("/post"),
			},
			status: 200,
		},
		{
			method: "GET",
			path:   "/http-routes/not-exist",
			token:  User1Token,
			schema: &errors.NotFoundError{},
			result: map[string]matcher{
				"$.message": contains("not found"),
			},
			status: 404,
		},

		// Delete

		{
			method: "DELETE",
			path:   "/http-routes/payload1/1",
			token:  User1Token,
			schema: actions.HTTPRoutesDeleteResult{},
			result: map[string]matcher{
				"$.path": equal("/get"),
			},
			status: 200,
		},
		{
			method: "DELETE",
			path:   "/http-routes/not-exist/1",
			token:  User1Token,
			schema: &errors.NotFoundError{},
			result: map[string]matcher{
				"$.message": contains("not found"),
			},
			status: 404,
		},
		{
			method: "DELETE",
			path:   "/http-routes/payload1/1337",
			token:  User1Token,
			schema: &errors.NotFoundError{},
			result: map[string]matcher{
				"$.message": contains("not found"),
			},
			status: 404,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s%s/%d", tt.method, tt.path, tt.status), func(t *testing.T) {
			setup(t)
			defer teardown(t)

			printers := make([]httpexpect.Printer, 0)

			if verbose {
				printers = append(printers, httpexpect.NewCurlPrinter(t))
				printers = append(printers, httpexpect.NewDebugPrinter(t, true))
			}

			cfg := httpexpect.Config{
				BaseURL:  srv.URL,
				Reporter: httpexpect.NewAssertReporter(t),
				Printers: printers,
			}

			if proxy != "" {
				proxyUrl, _ := url.Parse(proxy)
				cfg.Client = &http.Client{
					Transport: &http.Transport{
						Proxy: http.ProxyURL(proxyUrl),
					},
				}
			}
			e := httpexpect.WithConfig(cfg)

			req := e.Request(tt.method, tt.path)

			if tt.token != "" {
				req = req.WithHeader("Authorization", fmt.Sprintf("Bearer %s", tt.token))
			}

			if tt.json != "" {
				req = req.
					WithText(tt.json).
					WithHeader("Content-Type", "application/json; charset=utf-8")
			}

			if tt.query != "" {
				req = req.WithQueryString(tt.query)
			}

			schema, _ := jsonschema.Reflect(tt.schema).MarshalJSON()

			res := req.
				Expect().
				Status(tt.status).
				JSON().
				Schema(schema)

			for path, matcher := range tt.result {
				matcher(t, res.Path(path).Raw())
			}
		})
	}
}
