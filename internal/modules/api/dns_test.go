package api_test

import (
	"testing"

	"github.com/alecthomas/jsonschema"
	"github.com/bi-zone/sonar/internal/actions"
)

func TestDNSRecordsCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, User1Token)

	res := e.POST("/dns").
		WithJSON(map[string]interface{}{
			"payloadName": "payload1",
			"name":        "test",
			"type":        "a",
			"strategy":    "all",
			"values":      []string{"127.0.0.1"},
		}).
		Expect().
		Status(201).
		JSON()

	schema, _ := jsonschema.Reflect(&actions.DNSRecordsCreateResultData{}).MarshalJSON()

	res.Schema(schema)
}

func TestDNSRecordsCreate_BadRequest(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, User1Token)

	res := e.POST("/dns").
		WithJSON(map[string]interface{}{
			"invalid": "payload1",
			"name":    "test",
			"type":    "a",
		}).
		Expect().
		Status(400).
		JSON()

	res.Object().ContainsKey("message")
}

func TestDNSRecordsList_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, User1Token)

	res := e.GET("/dns/payload1").
		Expect().
		Status(200).
		JSON()

	schema, _ := jsonschema.Reflect(&actions.DNSRecordsListResultData{}).MarshalJSON()

	res.Schema(schema)
}

func TestDNSRecordsDelete_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, User1Token)

	res := e.DELETE("/dns/payload1/test-a/a").
		Expect().
		Status(200).
		JSON()

	schema, _ := jsonschema.Reflect(&actions.DNSRecordsDeleteResultData{}).MarshalJSON()

	res.Schema(schema)
}

func TestDNSRecordsDelete_NotFound(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)
	e = heAuth(e, User1Token)

	res := e.DELETE("/dns/payload1/not-exist/a").
		Expect().
		Status(404).
		JSON()

	res.Object().ContainsKey("message")
}
