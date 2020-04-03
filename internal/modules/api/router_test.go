package api_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/alecthomas/jsonschema"
)

type Payload struct {
	Name      string    `json:"name"`
	Subdomain string    `json:"subdomain"`
	CreatedAt time.Time `json:"createdAt"`
}

type Payloads []Payload

var (
	payloads, _ = jsonschema.Reflect(&Payloads{}).MarshalJSON()
	payload, _  = jsonschema.Reflect(&Payload{}).MarshalJSON()
)

func Test_listPayloads_Success(t *testing.T) {
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

func Test_createPayload_Success(t *testing.T) {
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

func Test_deletePayload_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, User1Token)

	_ = e.DELETE(fmt.Sprintf("/payloads/%s", "payload1")).
		Expect().
		Status(204).
		NoContent()

}

func Test_deletePayload_NotFound(t *testing.T) {
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
