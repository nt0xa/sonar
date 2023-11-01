package cmd

import (
	"github.com/spf13/cobra"

	"github.com/russtone/sonar/internal/utils/errors"
)

func (c *Command) withAuthCheck(cmd *cobra.Command) *cobra.Command {
	preRunE := cmd.PreRunE

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		_, err := c.actions.ProfileGet(cmd.Context())
		if err != nil {
			return err
		}

		if preRunE != nil {
			return preRunE(cmd, args)
		}

		return nil
	}

	return cmd
}

func (c *Command) withAdminCheck(cmd *cobra.Command) *cobra.Command {
	preRunE := cmd.PreRunE

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		profile, err := c.actions.ProfileGet(cmd.Context())
		if err != nil {
			return err
		}

		if !profile.IsAdmin {
			return errors.Forbidden()
		}

		if preRunE != nil {
			return preRunE(cmd, args)
		}

		return nil
	}

	return cmd
}
