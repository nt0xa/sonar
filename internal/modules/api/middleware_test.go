package api_test

import (
	"fmt"
	"testing"
)

func Test_checkAuth_NoToken(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)

	res := e.GET("/payloads").
		Expect().
		Status(401).
		JSON()

	res.Object().ContainsKey("message")
}

func Test_checkAuth_InvalidToken(t *testing.T) {
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

func Test_checkAuth_ValidToken(t *testing.T) {
	setup(t)
	defer teardown(t)

	e := heDefault(t)

	e.GET("/payloads").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", User1Token)).
		Expect().
		Status(200).
		JSON()
}
