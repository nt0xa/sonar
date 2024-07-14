package actions

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/utils/errors"
)

const (
	ProfileGetResultID = "profile/get"
)

type ProfileActions interface {
	ProfileGet(context.Context) (*ProfileGetResult, errors.Error)
}

type ProfileGetResult struct {
	User
}

func (r ProfileGetResult) ResultID() string {
	return ProfileGetResultID
}

func ProfileGetCommand(acts *Actions, local bool) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Get current user info",
	}

	return cmd, nil
}
