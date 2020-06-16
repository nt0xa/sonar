package cmd_test

import (
	"bytes"
	"context"
	"strings"

	"github.com/spf13/cobra"
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
		ID:   1,
		Name: "admin",
		Params: models.UserParams{
			Admin: true,
		},
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

func prepare() (*cobra.Command, *actions_mock.Actions, *ResultHandlerMock) {
	acts := &actions_mock.Actions{}
	hnd := &ResultHandlerMock{}
	return cmd.RootCmd(acts, hnd.Handle), acts, hnd
}

func execute(c *cobra.Command, user *models.User, args string) (string, error) {
	ctx := cmd.SetUser(context.Background(), user)

	c.SetArgs(strings.Split(args, " "))

	out := &bytes.Buffer{}

	c.SetOut(out)
	c.SetErr(out)

	err := c.ExecuteContext(ctx)

	return out.String(), err
}
