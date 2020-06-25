package cmd_test

import (
	"context"

	"github.com/stretchr/testify/mock"

	actions_mock "github.com/bi-zone/sonar/internal/actions/mock"
	"github.com/bi-zone/sonar/internal/cmd"
	"github.com/bi-zone/sonar/internal/models"
)

var (
	user = &models.User{
		ID:   1,
		Name: "test",
		Params: models.UserParams{
			TelegramID: 1337,
			APIToken:   "token",
		},
	}
	admin = &models.User{
		ID:      1,
		Name:    "admin",
		IsAdmin: true,
	}
	payload = &models.Payload{
		ID:     1,
		Name:   "payload",
		UserID: 1,
	}
	payloads = []*models.Payload{payload}
)

type ResultHandlerMock struct {
	mock.Mock
}

func (m *ResultHandlerMock) Handle(ctx context.Context, res interface{}) {
	m.Called(ctx, res)
}

func prepare() (*cmd.Command, *actions_mock.Actions, *ResultHandlerMock) {
	actions := &actions_mock.Actions{}
	handler := &ResultHandlerMock{}

	c := &cmd.Command{
		Actions:       actions,
		ResultHandler: handler.Handle,
	}

	return c, actions, handler
}
