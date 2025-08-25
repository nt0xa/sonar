package apiclient_test

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/actionsdb"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/modules/api"
	"github.com/nt0xa/sonar/internal/modules/api/apiclient"
	"github.com/nt0xa/sonar/internal/utils/errors"
	"github.com/nt0xa/sonar/pkg/telemetry"
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
	flag.StringVar(&proxy, "test.proxy", "", "Enables client proxy.")
	flag.Parse()
}

const (
	AdminToken = "a33bfdbfb3c62feb7ea59314dbd17426"
	UserToken  = "50c862e41d059eeca13adc7b276b46b7"
)

var (
	tf *testfixtures.Loader
	db *database.DB
	uc *apiclient.Client
	ac *apiclient.Client
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

	log := slog.New(slog.DiscardHandler)
	tel := telemetry.NewNoop()

	db, err = database.New(dsn, log, tel)
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
		testfixtures.Directory("../../../database/fixtures"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to load fixtures: %v", err)
		os.Exit(1)
	}

	api, err := api.New(&api.Config{Admin: AdminToken}, db, log, tel, nil, acts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to create api server: %v", err)
		os.Exit(1)
	}

	srv := httptest.NewServer(api.Router())

	var proxyURL *string

	if proxy != "" {
		proxyURL = &proxy
	}

	uc = apiclient.New(srv.URL, UserToken, true, proxyURL)
	ac = apiclient.New(srv.URL, AdminToken, true, proxyURL)

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

type matcher func(*testing.T, any)

func regex(re *regexp.Regexp) matcher {
	return func(t *testing.T, value any) {
		assert.Regexp(t, re, value)
	}
}

func withinDuration(d time.Duration) matcher {
	return func(t *testing.T, value any) {
		var (
			tm  time.Time
			err error
		)

		switch v := value.(type) {
		case string:
			tm, err = time.Parse(time.RFC3339, v)
		case time.Time:
			tm = v
		default:
			assert.True(t, false, "expected value %+v to be string or time.Time", v)
		}

		require.NoError(t, err)
		assert.WithinDuration(t, time.Now(), tm, d)
	}
}

func equal(expected any) matcher {
	return func(t *testing.T, value any) {
		assert.EqualValues(t, expected, value)
	}
}

//
// keyPath
//

func keyPath(obj any, kp string) (any, error) {
	keys := strings.Split(kp, ".")
	v := reflect.ValueOf(obj)

	for _, key := range keys {
		for v.Kind() == reflect.Pointer {
			v = v.Elem()
		}

		if v.Kind() != reflect.Struct &&
			v.Kind() != reflect.Array &&
			v.Kind() != reflect.Slice {
			return nil, fmt.Errorf("only accepts structs, arrays and slices; got %T", v)
		}

		i, err := strconv.ParseInt(key, 10, 32)
		if err == nil {
			v = v.Index(int(i))
		} else {
			v = v.FieldByName(key)
		}
	}

	return v.Interface(), nil
}

//
// Tests
//

func TestClient(t *testing.T) {
	tests := []struct {
		params any
		m      map[string]matcher
		err    errors.Error
	}{

		//
		// User
		//

		{
			nil,
			map[string]matcher{
				"Name": equal("user1"),
			},
			nil,
		},

		//
		// Payloads
		//

		// Create

		{
			actions.PayloadsCreateParams{
				Name:            "test",
				NotifyProtocols: []string{models.ProtoCategoryDNS.String()},
				StoreEvents:     true,
			},
			map[string]matcher{
				"Name":            equal("test"),
				"Subdomain":       regex(regexp.MustCompile("^[a-f0-9]{8}$")),
				"NotifyProtocols": equal([]string{"dns"}),
				"StoreEvents":     equal(true),
				"CreatedAt":       withinDuration(time.Second * 5),
			},
			nil,
		},

		// List

		{
			actions.PayloadsListParams{
				Name: "",
			},
			map[string]matcher{
				"0.Name": equal("test4"),
				"1.Name": equal("test3"),
			},
			nil,
		},

		// Update

		{
			actions.PayloadsUpdateParams{
				Name:            "payload1",
				NewName:         "test",
				NotifyProtocols: []string{models.ProtoCategoryHTTP.String()},
			},
			map[string]matcher{
				"Name":            equal("test"),
				"NotifyProtocols": equal([]string{"http"}),
			},
			nil,
		},

		// Delete

		{
			actions.PayloadsDeleteParams{
				Name: "payload1",
			},
			map[string]matcher{
				"Name": equal("payload1"),
			},
			nil,
		},

		// Clear

		{
			actions.PayloadsClearParams{
				Name: "1",
			},
			map[string]matcher{
				"0.Name": equal("payload1"),
			},
			nil,
		},

		//
		// DNS records
		//

		// Create

		{
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				Type:        models.DNSTypeA,
				Strategy:    models.DNSStrategyAll,
				Values:      []string{"10.1.1.2"},
				TTL:         100,
			},
			map[string]matcher{
				"Name":      equal("test"),
				"Type":      equal(models.DNSTypeA),
				"TTL":       equal(100),
				"Values":    equal([]string{"10.1.1.2"}),
				"CreatedAt": withinDuration(time.Second * 5),
			},
			nil,
		},

		// List

		{
			actions.DNSRecordsListParams{
				PayloadName: "payload1",
			},
			map[string]matcher{
				"0.Name": equal("test-a"),
				"8.Name": equal("test-rebind"),
			},
			nil,
		},

		// Delete

		{
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload1",
				Index:       1,
			},
			map[string]matcher{
				"Name": equal("test-a"),
			},
			nil,
		},

		// Clear

		{
			actions.DNSRecordsClearParams{
				PayloadName: "payload1",
			},
			map[string]matcher{
				"0.Name": equal("test-a"),
				"8.Name": equal("test-rebind"),
			},
			nil,
		},

		//
		// Users
		//

		// Create

		{
			actions.UsersCreateParams{
				Name: "test",
				Params: models.UserParams{
					TelegramID: 1234,
					APIToken:   "test",
				},
				IsAdmin: false,
			},
			map[string]matcher{
				"Name":              equal("test"),
				"Params.TelegramID": equal(1234),
				"Params.APIToken":   equal("test"),
				"IsAdmin":           equal(false),
				"CreatedAt":         withinDuration(time.Second * 5),
			},
			nil,
		},

		// Delete

		{
			actions.UsersDeleteParams{
				Name: "user1",
			},
			map[string]matcher{
				"Name": equal("user1"),
			},
			nil,
		},

		//
		// Events
		//

		// List

		{
			actions.EventsListParams{
				PayloadName: "payload1",
			},
			map[string]matcher{
				"0.Protocol": equal("http"),
				"9.Protocol": equal("dns"),
			},
			nil,
		},

		// Get

		{
			actions.EventsGetParams{
				PayloadName: "payload1",
				Index:       2,
			},
			map[string]matcher{
				"Protocol": equal("http"),
			},
			nil,
		},

		//
		// HTTP routes
		//

		// Create

		{
			actions.HTTPRoutesCreateParams{
				PayloadName: "payload1",
				Method:      "PUT",
				Path:        "/123",
				Code:        302,
				Headers: map[string][]string{
					"Location": {"http://example.com"},
				},
				IsDynamic: true,
			},
			map[string]matcher{
				"Method":    equal("PUT"),
				"Path":      equal("/123"),
				"Code":      equal(302),
				"Headers":   equal(map[string][]string{"Location": {"http://example.com"}}),
				"IsDynamic": equal(true),
				"CreatedAt": withinDuration(time.Second * 5),
			},
			nil,
		},

		// Update

		{
			actions.HTTPRoutesUpdateParams{
				Payload: "payload1",
				Index:   1,
				Method:  ptr("PUT"),
				Path:    ptr("/123"),
				Code:    ptr(302),
				Headers: map[string][]string{
					"Location": {"http://example.com"},
				},
				IsDynamic: ptr(true),
				Body:      ptr("dGVzdA=="),
			},
			map[string]matcher{
				"Method":    equal("PUT"),
				"Path":      equal("/123"),
				"Code":      equal(302),
				"Headers":   equal(map[string][]string{"Location": {"http://example.com"}}),
				"IsDynamic": equal(true),
				"Body":      equal("dGVzdA=="),
			},
			nil,
		},

		// List

		{
			actions.HTTPRoutesListParams{
				PayloadName: "payload1",
			},
			map[string]matcher{
				"0.Path": equal("/get"),
				"3.Path": equal("/redirect"),
			},
			nil,
		},

		// Delete

		{
			actions.HTTPRoutesDeleteParams{
				PayloadName: "payload1",
				Index:       1,
			},
			map[string]matcher{
				"Path": equal("/get"),
			},
			nil,
		},

		// Clear

		{
			actions.HTTPRoutesClearParams{
				PayloadName: "payload1",
			},
			map[string]matcher{
				"0.Path": equal("/get"),
				"4.Path": equal("/route/{test}"),
			},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T", tt.params), func(t *testing.T) {
			setup(t)
			defer teardown(t)

			var (
				res any
				err errors.Error
			)

			switch p := tt.params.(type) {

			// Payloads
			case actions.PayloadsCreateParams:
				res, err = uc.PayloadsCreate(context.Background(), p)
			case actions.PayloadsListParams:
				res, err = uc.PayloadsList(context.Background(), p)
			case actions.PayloadsUpdateParams:
				res, err = uc.PayloadsUpdate(context.Background(), p)
			case actions.PayloadsDeleteParams:
				res, err = uc.PayloadsDelete(context.Background(), p)
			case actions.PayloadsClearParams:
				res, err = uc.PayloadsClear(context.Background(), p)

				// DNS records
			case actions.DNSRecordsCreateParams:
				res, err = uc.DNSRecordsCreate(context.Background(), p)
			case actions.DNSRecordsListParams:
				res, err = uc.DNSRecordsList(context.Background(), p)
			case actions.DNSRecordsDeleteParams:
				res, err = uc.DNSRecordsDelete(context.Background(), p)
			case actions.DNSRecordsClearParams:
				res, err = uc.DNSRecordsClear(context.Background(), p)

			// Events
			case actions.EventsListParams:
				res, err = uc.EventsList(context.Background(), p)
			case actions.EventsGetParams:
				res, err = uc.EventsGet(context.Background(), p)

			// Users
			case actions.UsersCreateParams:
				res, err = ac.UsersCreate(context.Background(), p)
			case actions.UsersDeleteParams:
				res, err = ac.UsersDelete(context.Background(), p)

			// HTTP routes
			case actions.HTTPRoutesCreateParams:
				res, err = uc.HTTPRoutesCreate(context.Background(), p)
			case actions.HTTPRoutesListParams:
				res, err = uc.HTTPRoutesList(context.Background(), p)
			case actions.HTTPRoutesDeleteParams:
				res, err = uc.HTTPRoutesDelete(context.Background(), p)
			case actions.HTTPRoutesClearParams:
				res, err = uc.HTTPRoutesClear(context.Background(), p)
			case actions.HTTPRoutesUpdateParams:
				res, err = uc.HTTPRoutesUpdate(context.Background(), p)

			// Profile
			case nil:
				res, err = uc.ProfileGet(context.Background())

			default:
				panic("not implemented: add new case to switch")
			}

			if tt.err != nil {
				require.Error(t, err)
				assert.IsType(t, tt.err, err)
			} else {
				require.NoError(t, err)

				for kp, matcher := range tt.m {
					v, err := keyPath(res, kp)
					require.NoError(t, err)
					matcher(t, v)
				}
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
