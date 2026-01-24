package database_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/database"
)

func TestHTTPRoutesCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.HTTPRoutesCreate(t.Context(), database.HTTPRoutesCreateParams{
		PayloadID: 1,
		Method:    "GET",
		Path:      "/test",
		Code:      200,
		Headers: map[string][]string{
			"Test": {"test"},
		},
		Body: []byte("body"),
	})
	assert.NoError(t, err)
	assert.NotZero(t, o.ID)
	assert.WithinDuration(t, time.Now(), o.CreatedAt, 5*time.Second)
	assert.EqualValues(t, 6, o.Index)
}

func TestHTTPRoutesCreate_Duplicate(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.HTTPRoutesCreate(t.Context(), database.HTTPRoutesCreateParams{
		PayloadID: 1,
		Method:    "GET",
		Path:      "/get",
		Code:      200,
		Headers: map[string][]string{
			"Test": {"test"},
		},
		Body: []byte("body"),
	})
	assert.Error(t, err)
}

func TestHTTPRoutesGetByID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.HTTPRoutesGetByID(t.Context(), 2)
	assert.NoError(t, err)
	assert.Equal(t, "/post", o.Path)
}

func TestHTTPRoutesGetByID_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.HTTPRoutesGetByID(t.Context(), 1337)
	assert.Error(t, err)
	assert.Nil(t, o)
	assert.Error(t, err, sql.ErrNoRows.Error())
}

func TestHTTPRoutesDelete_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	err := db.HTTPRoutesDelete(t.Context(), 1)
	assert.NoError(t, err)
}

func TestHTTPRoutesUpdate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.HTTPRoutesGetByID(t.Context(), 1)
	require.NoError(t, err)
	assert.NotNil(t, o)

	updated, err := db.HTTPRoutesUpdate(t.Context(), database.HTTPRoutesUpdateParams{
		ID:        o.ID,
		PayloadID: o.PayloadID,
		Method:    "HEAD",
		Path:      "/updated",
		Code:      o.Code,
		Headers:   o.Headers,
		Body:      o.Body,
		IsDynamic: o.IsDynamic,
	})
	require.NoError(t, err)

	o2, err := db.HTTPRoutesGetByID(t.Context(), 1)
	require.NoError(t, err)
	assert.Equal(t, updated.Method, o2.Method)
	assert.Equal(t, updated.Path, o2.Path)
}

func TestHTTPRoutesGetByPayloadID(t *testing.T) {
	setup(t)
	defer teardown(t)

	l, err := db.HTTPRoutesGetByPayloadID(t.Context(), 1)
	assert.NoError(t, err)
	assert.Len(t, l, 5)
	assert.EqualValues(t, 1, l[0].Index)
	assert.EqualValues(t, 5, l[len(l)-1].Index)
}

func TestHTTPRoutesGetByPayloadIDAndIndex(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.HTTPRoutesGetByPayloadIDAndIndex(t.Context(), 1, 3)
	assert.NoError(t, err)
	assert.EqualValues(t, "/delete", o.Path)

	// Not exist
	_, err = db.HTTPRoutesGetByPayloadIDAndIndex(t.Context(), 1, 1337)
	assert.Error(t, err)
}

func TestHTTPRoutesGetByPayloadMethodAndPath(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.HTTPRoutesGetByPayloadMethodAndPath(t.Context(), 1, "POST", "/post")
	assert.NoError(t, err)
	assert.EqualValues(t, "/post", o.Path)

	// Not exist
	_, err = db.HTTPRoutesGetByPayloadMethodAndPath(t.Context(), 1337, "POST", "/post")
	assert.Error(t, err)
}

func TestHTTPRoutesDeleteAllByPayloadID(t *testing.T) {
	setup(t)
	defer teardown(t)

	l, err := db.HTTPRoutesDeleteAllByPayloadID(t.Context(), 1)
	assert.NoError(t, err)
	assert.Len(t, l, 5)
}

func TestHTTPRoutesDeleteAllByPayloadIDAndName(t *testing.T) {
	setup(t)
	defer teardown(t)

	l, err := db.HTTPRoutesDeleteAllByPayloadIDAndPath(t.Context(), 1, "/get")
	assert.NoError(t, err)
	assert.Len(t, l, 1)
}
