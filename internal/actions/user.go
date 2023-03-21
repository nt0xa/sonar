package actions

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/russtone/sonar/internal/utils/errors"
)

type ProfileActions interface {
	ProfileGet(context.Context) (*ProfileGetResult, errors.Error)
}

type ProfileGetResult struct {
	User
}

func (r ProfileGetResult) ResultID() string {
	return "profile/get"
}

func ProfileGetCommand(local bool) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Get current user info",
	}

	return cmd, nil
}
