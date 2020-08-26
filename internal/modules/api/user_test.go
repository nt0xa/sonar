package api_test

import (
	"testing"

	"github.com/alecthomas/jsonschema"
	"github.com/bi-zone/sonar/internal/actions"
)

func TestUserCurrent_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, User1Token)

	res := e.GET("/user").
		Expect().
		Status(200).
		JSON()

	schema, _ := jsonschema.Reflect(&actions.User{}).MarshalJSON()

	res.Schema(schema)
}
