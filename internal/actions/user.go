package actions

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/utils/errors"
)

type UserActions interface {
	UserCurrent(context.Context) (UserCurrentResult, errors.Error)
}

type UserHandler interface {
	UserCurrent(context.Context, UserCurrentResult)
}

type UserCurrentResult *User

func UserCurrentCommand() (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Get current user info",
	}

	return cmd, nil
}
