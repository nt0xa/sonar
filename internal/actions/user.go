package actions

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/russtone/sonar/internal/utils/errors"
)

type ProfileActions interface {
	ProfileGet(context.Context) (ProfileGetResult, errors.Error)
}

type ProfileHandler interface {
	ProfileGet(context.Context, ProfileGetResult)
}

type ProfileGetResult *User

func ProfileGetCommand(local bool) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Get current user info",
	}

	return cmd, nil
}
