package apiclient_test

import (
	"context"
	"flag"
	"fmt"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-testfixtures/testfixtures"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/modules/api"
	"github.com/bi-zone/sonar/internal/modules/api/apiclient"
	"github.com/bi-zone/sonar/internal/testutils"
	"github.com/bi-zone/sonar/internal/utils/errors"
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
	db   *database.DB
	tf   *testfixtures.Context
	acts actions.Actions
	srv  *httptest.Server
	uc   *apiclient.Client
	ac   *apiclient.Client

	log = logrus.New()

	g = testutils.Globals(
		testutils.DB(&database.Config{
			DSN:        os.Getenv("SONAR_DB_DSN"),
			Migrations: "../../../database/migrations",
		}, &db),
		testutils.Fixtures(&db, "../../../database/fixtures", &tf),
		testutils.ActionsDB(&db, log, &acts),
		testutils.APIServer(&api.Config{Admin: AdminToken}, &db, log, &acts, &srv),
		testutils.APIClient(&srv, UserToken, &uc),
		testutils.APIClient(&srv, AdminToken, &ac),
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

func equal(expected interface{}) matcher {
	return func(t *testing.T, value interface{}) {
		assert.EqualValues(t, expected, value)
	}
}

func contains(s interface{}) matcher {
	return func(t *testing.T, value interface{}) {
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
// keyPath
//

func keyPath(obj interface{}, kp string) (interface{}, error) {
	keys := strings.Split(kp, ".")
	v := reflect.ValueOf(obj)

	for _, key := range keys {
		for v.Kind() == reflect.Ptr {
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
		params interface{}
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
			},
			map[string]matcher{
				"Name":            equal("test"),
				"Subdomain":       regex(regexp.MustCompile("^[a-f0-9]{8}$")),
				"NotifyProtocols": equal([]string{"dns"}),
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
				"0.Name": equal("payload1"),
				"1.Name": equal("payload4"),
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
				"Name":      equal("test.c1da9f3d"),
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
				"0.Name": equal("test-a.c1da9f3d"),
				"8.Name": equal("test-rebind.c1da9f3d"),
			},
			nil,
		},

		// Delete

		{
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload1",
				Name:        "test-a",
				Type:        "a",
			},
			map[string]matcher{
				"Name": equal("test-a.c1da9f3d"),
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
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T", tt.params), func(t *testing.T) {
			setup(t)
			defer teardown(t)

			var (
				res interface{}
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

				// DNS records
			case actions.DNSRecordsCreateParams:
				res, err = uc.DNSRecordsCreate(context.Background(), p)
			case actions.DNSRecordsListParams:
				res, err = uc.DNSRecordsList(context.Background(), p)
			case actions.DNSRecordsDeleteParams:
				res, err = uc.DNSRecordsDelete(context.Background(), p)

			// Users
			case actions.UsersCreateParams:
				res, err = ac.UsersCreate(context.Background(), p)
			case actions.UsersDeleteParams:
				res, err = ac.UsersDelete(context.Background(), p)

			// User
			default:
				res, err = uc.UserCurrent(context.Background())
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
