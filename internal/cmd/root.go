package cmd

import (
	"context"

	"github.com/Masterminds/sprig"
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/actions"
)

func init() {
	cobra.AddTemplateFuncs(sprig.TxtFuncMap())
}

type ResultHandler func(context.Context, interface{})

func RootCmd(acts actions.Actions, handler ResultHandler) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sonarctl",
		Short: "CLI to control your sonar server",
	}

	cmd.AddCommand(CreatePayloadCmd(acts, handler))
	cmd.AddCommand(DeletePayloadCmd(acts, handler))
	cmd.AddCommand(ListPayloadCmd(acts, handler))
	cmd.AddCommand(UsersCmd(acts, handler))

	return cmd
}
