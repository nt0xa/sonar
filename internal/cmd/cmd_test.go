package cmd_test

import (
	"context"

	actions_mock "github.com/bi-zone/sonar/internal/actions/mock"
	"github.com/bi-zone/sonar/internal/cmd"
)

var (
	ctx = context.WithValue(context.Background(), "key", "value")
)

func prepare() (cmd.Command, *actions_mock.Actions, *actions_mock.ResultHandler) {
	actions := &actions_mock.Actions{}
	handler := &actions_mock.ResultHandler{}

	c := cmd.New(actions, handler, nil)

	return c, actions, handler
}
