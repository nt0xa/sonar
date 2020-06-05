package cmd

import (
	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/spf13/cobra"
)

type ResultHandler func(*database.User, interface{})

func RootCmd(actions actions.Actions, handler ResultHandler) *cobra.Command {
	var root = &cobra.Command{
		Use:   "sonarctl",
		Short: "CLI to control your sonar server",
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}

	root.AddCommand(CreatePayloadCmd(actions, handler))

	return root
}
