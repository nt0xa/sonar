package api_test

import (
	"fmt"
	"testing"
)

func TestAuth_NoToken(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)

	res := e.GET("/payloads").
		Expect().
		Status(401).
		JSON()

	res.Object().ContainsKey("message")
}

func TestAuth_InvalidToken(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)

	res := e.GET("/payloads").
		WithHeader("Authorization", "Bearer invalid").
		Expect().
		Status(401).
		JSON()

	res.Object().ContainsKey("message")
}

func TestAuth_ValidToken(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)

	e.GET("/payloads").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", User1Token)).
		Expect().
		Status(200).
		JSON()
}

func TestAdmin_Forbidden(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)

	res := e.POST("/users").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", User1Token)).
		Expect().
		Status(403).
		JSON()

	res.Object().ContainsKey("message")
}

func TestAdmin_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)

	_ = e.POST("/users").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", AdminToken)).
		WithJSON(map[string]interface{}{
			"name": "test",
		}).
		Expect().
		Status(201).
		JSON()
}
