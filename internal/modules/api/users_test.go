package api_test

import (
	"fmt"
	"testing"

	"github.com/alecthomas/jsonschema"
	"github.com/bi-zone/sonar/internal/actions"
)

type User = actions.User

var (
	user, _ = jsonschema.Reflect(&User{}).MarshalJSON()
)

func TestUsersCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, AdminToken)

	res := e.POST("/users").
		WithJSON(map[string]interface{}{
			"name": "test",
		}).
		Expect().
		Status(201).
		JSON()

	res.Schema(user)
}

func TestUsersCreate_BadRequest(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, AdminToken)

	res := e.POST("/users").
		WithJSON(map[string]interface{}{
			"invalid": "test",
		}).
		Expect().
		Status(400).
		JSON()

	res.Object().ContainsKey("message")
}

func TestUsersCreate_Conflict(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, AdminToken)

	res := e.POST("/users").
		WithJSON(map[string]interface{}{
			"name": "user1",
		}).
		Expect().
		Status(409).
		JSON()

	res.Object().ContainsKey("message")
}

func TestUsersDelete_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, AdminToken)

	res := e.DELETE(fmt.Sprintf("/users/%s", "user2")).
		Expect().
		Status(200).
		JSON()

	res.Schema(user)
}

func TestUsersDelete_NotFound(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, AdminToken)

	res := e.DELETE(fmt.Sprintf("/users/%s", "invalid")).
		Expect().
		Status(404).
		JSON()

	res.Object().ContainsKey("message")
}
