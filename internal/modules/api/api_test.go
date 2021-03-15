package api_test

import (
	"flag"
	"fmt"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/alecthomas/jsonschema"
	"github.com/gavv/httpexpect"
	"github.com/go-testfixtures/testfixtures"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/modules/api"
	"github.com/bi-zone/sonar/internal/testutils"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

// Flags
var (
	logs    bool
	verbose bool
)

var _ = func() bool {
	testing.Init()
	return true
}()

func init() {
	flag.BoolVar(&logs, "test.logs", false, "Enables logger output.")
	flag.BoolVar(&verbose, "test.verbose", false, "Enables verbose HTTP printing.")
	flag.Parse()
}

const (
	AdminToken = "a33bfdbfb3c62feb7ea59314dbd17426"
	User1Token = "50c862e41d059eeca13adc7b276b46b7"
	User2Token = "7001f2d819d3d5fb0b1fd75dd38eb34e"
)

var (
	db   *database.DB
	tf   *testfixtures.Context
	acts actions.Actions
	srv  *httptest.Server

	log = logrus.New()

	g = testutils.Globals(
		testutils.DB(&database.Config{
			DSN:        os.Getenv("SONAR_DB_DSN"),
			Migrations: "../../database/migrations",
		}, &db),
		testutils.Fixtures(&db, "../../database/fixtures", &tf),
		testutils.ActionsDB(&db, log, &acts),
		testutils.APIServer(&api.Config{Admin: AdminToken}, &db, log, &acts, &srv),
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
		assert.WithinDuration(t, time.Now().UTC(), tt, d)
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
			json:   `{"name": "test", "notifyProtocols": ["dns", "smtp"]}`,
			schema: (actions.PayloadsCreateResult)(nil),
			result: map[string]matcher{
				"$.subdomain": regex(regexp.MustCompile("^[a-f0-9]{8}$")),
				"$.name":      equal("test"),
				"$.notifyProtocols": equal(
					[]interface{}{
						models.ProtoCategoryDNS.String(),
						models.ProtoCategorySMTP.String(),
					},
				),
				"$.createdAt": withinDuration(time.Second * 10),
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
			json:   `{"name":"test", "notifyProtocols": ["smtp"]}`,
			schema: (actions.PayloadsUpdateResult)(nil),
			result: map[string]matcher{
				"$.name":            equal("test"),
				"$.notifyProtocols": equal([]interface{}{models.ProtoCategorySMTP.String()}),
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
			schema: (actions.PayloadsDeleteResult)(nil),
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

		//
		// DNS records
		//

		// Create

		{
			method: "POST",
			path:   "/dnsrecords",
			token:  User1Token,
			json:   `{"payloadName": "payload1", "name": "test", "type": "a", "ttl": 100, "values": ["127.0.0.1"], "strategy": "all"}`,
			schema: (actions.DNSRecordsCreateResult)(nil),
			result: map[string]matcher{
				"$.record.name":      equal("test"),
				"$.record.type":      equal("A"),
				"$.record.ttl":       equal(100),
				"$.record.strategy":  equal("all"),
				"$.record.values":    equal([]interface{}{"127.0.0.1"}),
				"$.record.createdAt": withinDuration(time.Second * 10),
			},
			status: 201,
		},
		{
			method: "POST",
			path:   "/dnsrecords",
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
			path:   "/dnsrecords",
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
			path:   "/dnsrecords",
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
			path:   "/dnsrecords/payload1",
			token:  User1Token,
			schema: (actions.DNSRecordsListResult)(nil),
			result: map[string]matcher{
				"$.records":         length(9),
				"$.records[0].name": equal("test-a"),
				"$.records[1].name": equal("test-aaaa"),
			},
			status: 200,
		},
		{
			method: "GET",
			path:   "/dnsrecords/not-exist",
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
			path:   "/dnsrecords/payload1/test-a/a",
			token:  User1Token,
			schema: (actions.DNSRecordsDeleteResult)(nil),
			result: map[string]matcher{
				"$.record.name": equal("test-a"),
			},
			status: 200,
		},
		{
			method: "DELETE",
			path:   "/dnsrecords/not-exist/test-a/a",
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
			path:   "/user",
			token:  User1Token,
			schema: (actions.UserCurrentResult)(nil),
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
			schema: (actions.UsersCreateResult)(nil),
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
			schema: (actions.UsersCreateResult)(nil),
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

			e := httpexpect.WithConfig(httpexpect.Config{
				BaseURL:  srv.URL,
				Reporter: httpexpect.NewAssertReporter(t),
				Printers: printers,
			})

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
