package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/actions"
)

func CreatePayloadCmd(acts actions.Actions, handler ResultHandler) *cobra.Command {
	p := actions.CreatePayloadParams{}

	cmd := &cobra.Command{
		Use:   "new NAME",
		Short: "Create new payload",
		Args:  cobra.ExactArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			p.Name = args[0]
			return p.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			u, err := GetUser(cmd.Context())
			if err != nil {
				return err
			}

			res, err := acts.CreatePayload(u, p)
			if err != nil {
				return err
			}

			handler(u, res)

			return nil
		},
	}

	return cmd
}
