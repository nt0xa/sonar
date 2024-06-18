package actionsdb_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/actionsdb"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/utils/errors"
)

func TestHTTPRoutesCreate_Success(t *testing.T) {

	tests := []struct {
		name string
		p    actions.HTTPRoutesCreateParams
	}{
		{
			"GET",
			actions.HTTPRoutesCreateParams{
				PayloadName: "payload1",
				Method:      "GET",
				Path:        "/test",
				Code:        200,
				Headers: models.Headers{
					"Test": {"test"},
				},
				Body:      "test",
				IsDynamic: false,
			},
		},
		{
			"POST",
			actions.HTTPRoutesCreateParams{
				PayloadName: "payload1",
				Method:      "POST",
				Path:        "/test-2",
				Code:        201,
				Headers: models.Headers{
					"Test": {"test"},
				},
				Body:      "test",
				IsDynamic: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			u, err := db.UsersGetByID(1)
			require.NoError(t, err)

			ctx := actionsdb.SetUser(context.Background(), u)

			r, err := acts.HTTPRoutesCreate(ctx, tt.p)
			require.NoError(t, err)
			require.NotNil(t, r)

			assert.Equal(t, tt.p.Method, r.Method)
			assert.Equal(t, tt.p.Path, r.Path)
			assert.Equal(t, tt.p.Code, r.Code)
			assert.Equal(t, tt.p.Headers, r.Headers)
			assert.Equal(t, tt.p.Body, r.Body)
			assert.Equal(t, tt.p.IsDynamic, r.IsDynamic)
		})
	}
}

func TestHTTPRoutesCreate_Error(t *testing.T) {

	tests := []struct {
		name   string
		userID int
		p      actions.HTTPRoutesCreateParams
		err    errors.Error
	}{
		{
			"no user in ctx",
			0,
			actions.HTTPRoutesCreateParams{
				PayloadName: "payload1",
				Method:      "GET",
				Path:        "/test",
				Code:        200,
				Headers: models.Headers{
					"Test": {"test"},
				},
				Body:      "test",
				IsDynamic: false,
			},
			&errors.InternalError{},
		},
		{
			"empty path",
			1,
			actions.HTTPRoutesCreateParams{
				PayloadName: "payload1",
				Method:      "GET",
				Path:        "",
				Code:        200,
				Headers: models.Headers{
					"Test": {"test"},
				},
				Body:      "test",
				IsDynamic: false,
			},
			&errors.ValidationError{},
		},
		{
			"empty payload name",
			1,
			actions.HTTPRoutesCreateParams{
				PayloadName: "",
				Method:      "GET",
				Path:        "/test",
				Code:        200,
				Headers: models.Headers{
					"Test": {"test"},
				},
				Body:      "test",
				IsDynamic: false,
			},
			&errors.ValidationError{},
		},
		{
			"invalid body",
			1,
			actions.HTTPRoutesCreateParams{
				PayloadName: "payload1",
				Method:      "GET",
				Path:        "/test",
				Code:        200,
				Headers: models.Headers{
					"Test": {"test"},
				},
				Body:      "xxxxxxx",
				IsDynamic: false,
			},
			&errors.ValidationError{},
		},
		{
			"not existing payload name",
			1,
			actions.HTTPRoutesCreateParams{
				PayloadName: "not-exist",
				Method:      "GET",
				Path:        "/test",
				Code:        200,
				Headers: models.Headers{
					"Test": {"test"},
				},
				Body:      "test",
				IsDynamic: false,
			},
			&errors.NotFoundError{},
		},
		{
			"duplicate name and type",
			1,
			actions.HTTPRoutesCreateParams{
				PayloadName: "payload1",
				Method:      "GET",
				Path:        "/get",
				Code:        200,
				Headers: models.Headers{
					"Test": {"test"},
				},
				Body:      "test",
				IsDynamic: false,
			},
			&errors.ConflictError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			ctx := context.Background()

			if tt.userID != 0 {
				u, err := db.UsersGetByID(1)
				require.NoError(t, err)

				ctx = actionsdb.SetUser(context.Background(), u)
			}

			_, err := acts.HTTPRoutesCreate(ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestHTTPRoutesDelete_Success(t *testing.T) {

	tests := []struct {
		method string
		path   string
		p      actions.HTTPRoutesDeleteParams
	}{
		{
			"GET",
			"/get",
			actions.HTTPRoutesDeleteParams{
				PayloadName: "payload1",
				Index:       1,
			},
		},
		{
			"POST",
			"/post",
			actions.HTTPRoutesDeleteParams{
				PayloadName: "payload1",
				Index:       2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			u, err := db.UsersGetByID(1)
			require.NoError(t, err)

			ctx := actionsdb.SetUser(context.Background(), u)

			_, err = acts.HTTPRoutesDelete(ctx, tt.p)
			assert.NoError(t, err)

			p, err := db.PayloadsGetByUserAndName(u.ID, tt.p.PayloadName)
			assert.NoError(t, err)

			_, err = db.HTTPRoutesGetByPayloadMethodAndPath(p.ID, tt.method, tt.path)
			assert.Error(t, err, sql.ErrNoRows)
		})
	}
}

func TestHTTPRoutesDelete_Error(t *testing.T) {
	tests := []struct {
		name   string
		userID int
		p      actions.HTTPRoutesDeleteParams
		err    errors.Error
	}{
		{
			"no user in ctx",
			0,
			actions.HTTPRoutesDeleteParams{
				PayloadName: "payload1",
				Index:       1,
			},
			&errors.InternalError{},
		},
		{
			"empty payload name",
			1,
			actions.HTTPRoutesDeleteParams{
				PayloadName: "",
				Index:       1,
			},
			&errors.ValidationError{},
		},
		{
			"not existing payload",
			1,
			actions.HTTPRoutesDeleteParams{
				PayloadName: "not-exist",
				Index:       1,
			},
			&errors.NotFoundError{},
		},
		{
			"not existing index",
			1,
			actions.HTTPRoutesDeleteParams{
				PayloadName: "payload1",
				Index:       1337,
			},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			ctx := context.Background()

			if tt.userID != 0 {
				u, err := db.UsersGetByID(1)
				require.NoError(t, err)

				ctx = actionsdb.SetUser(context.Background(), u)
			}

			_, err := acts.HTTPRoutesDelete(ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestHTTPRoutesList_Success(t *testing.T) {

	tests := []struct {
		name  string
		p     actions.HTTPRoutesListParams
		count int
	}{
		{
			"payload1",
			actions.HTTPRoutesListParams{
				PayloadName: "payload1",
			},
			5,
		},
		{
			"payload4",
			actions.HTTPRoutesListParams{
				PayloadName: "payload4",
			},
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			u, err := db.UsersGetByID(1)
			require.NoError(t, err)

			ctx := actionsdb.SetUser(context.Background(), u)

			list, err := acts.HTTPRoutesList(ctx, tt.p)
			assert.NoError(t, err)
			assert.Len(t, list, tt.count)
		})
	}
}

func TestHTTPRoutesList_Error(t *testing.T) {

	tests := []struct {
		name   string
		userID int
		p      actions.HTTPRoutesListParams
		err    errors.Error
	}{
		{
			"no user in ctx",
			0,
			actions.HTTPRoutesListParams{
				PayloadName: "payload1",
			},
			&errors.InternalError{},
		},
		{
			"empty payload name",
			1,
			actions.HTTPRoutesListParams{
				PayloadName: "",
			},
			&errors.ValidationError{},
		},
		{
			"not existing payload",
			1,
			actions.HTTPRoutesListParams{
				PayloadName: "not-exist",
			},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			ctx := context.Background()

			if tt.userID != 0 {
				u, err := db.UsersGetByID(1)
				require.NoError(t, err)

				ctx = actionsdb.SetUser(context.Background(), u)
			}

			_, err := acts.HTTPRoutesList(ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestHTTPRoutesClear_Success(t *testing.T) {

	tests := []struct {
		name  string
		p     actions.HTTPRoutesClearParams
		count int
	}{
		{
			"payload1",
			actions.HTTPRoutesClearParams{
				PayloadName: "payload1",
			},
			5,
		},
		{
			"payload1",
			actions.HTTPRoutesClearParams{
				PayloadName: "payload1",
				Path:        "/post",
			},
			1,
		},
		{
			"payload4",
			actions.HTTPRoutesClearParams{
				PayloadName: "payload4",
			},
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			u, err := db.UsersGetByID(1)
			require.NoError(t, err)

			ctx := actionsdb.SetUser(context.Background(), u)

			list, err := acts.HTTPRoutesClear(ctx, tt.p)
			assert.NoError(t, err)
			assert.Len(t, list, tt.count)
		})
	}
}

func TestHTTPRoutesClear_Error(t *testing.T) {

	tests := []struct {
		name   string
		userID int
		p      actions.HTTPRoutesClearParams
		err    error
	}{
		{
			"no user in ctx",
			0,
			actions.HTTPRoutesClearParams{
				Path: "/get",
			},
			&errors.InternalError{},
		},
		{
			"not existing payload",
			1,
			actions.HTTPRoutesClearParams{
				PayloadName: "not-exist",
			},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			ctx := context.Background()
			if tt.userID != 0 {
				u, err := db.UsersGetByID(1)
				require.NoError(t, err)

				ctx = actionsdb.SetUser(context.Background(), u)
			}

			_, err := acts.HTTPRoutesClear(ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}
