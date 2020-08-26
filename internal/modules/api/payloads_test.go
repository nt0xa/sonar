package api_test

import (
	"fmt"
	"testing"

	"github.com/alecthomas/jsonschema"
	"github.com/bi-zone/sonar/internal/actions"
)

type Payload = actions.Payload
type Payloads = []Payload

var (
	payloads, _ = jsonschema.Reflect(&Payloads{}).MarshalJSON()
	payload, _  = jsonschema.Reflect(&Payload{}).MarshalJSON()
)

func TestPayloadsCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, User1Token)

	res := e.POST("/payloads").
		WithJSON(map[string]interface{}{
			"name": "test",
		}).
		Expect().
		Status(201).
		JSON()

	res.Schema(payload)
}

func TestPayloadsCreate_BadRequest(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, User1Token)

	res := e.POST("/payloads").
		WithJSON(map[string]interface{}{
			"invalid": "test",
		}).
		Expect().
		Status(400).
		JSON()

	res.Object().ContainsKey("message")
}

func TestPayloadsList_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, User1Token)

	res := e.GET("/payloads").
		Expect().
		Status(200).
		JSON()

	res.Schema(payloads)
}

func TestPayloadsUpdate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, User1Token)

	res := e.PUT("/payloads/payload1").
		WithJSON(map[string]interface{}{
			"name": "test",
		}).
		Expect().
		Status(200).
		JSON()

	res.Schema(payload)
}

func TestPayloadsDelete_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, User1Token)

	res := e.DELETE(fmt.Sprintf("/payloads/%s", "payload1")).
		Expect().
		Status(200).
		JSON()

	res.Schema(payload)
}

func TestPayloadsDelete_NotFound(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, User1Token)

	res := e.DELETE(fmt.Sprintf("/payloads/%s", "not-exist")).
		Expect().
		Status(404).
		JSON()

	res.Object().ContainsKey("message")
}
