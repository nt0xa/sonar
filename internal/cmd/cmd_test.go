package cmd_test

import (
	"context"

	"github.com/stretchr/testify/mock"

	actions_mock "github.com/bi-zone/sonar/internal/actions/mock"
	"github.com/bi-zone/sonar/internal/cmd"
)

var (
	ctx = context.WithValue(context.Background(), "key", "value")
)

type ResultHandlerMock struct {
	mock.Mock
}

func (m *ResultHandlerMock) Handle(ctx context.Context, res interface{}) {
	m.Called(ctx, res)
}

func prepare() (cmd.Command, *actions_mock.Actions, *ResultHandlerMock) {
	actions := &actions_mock.Actions{}
	handler := &ResultHandlerMock{}

	c := cmd.New(actions, handler.Handle, nil)

	return c, actions, handler
}
