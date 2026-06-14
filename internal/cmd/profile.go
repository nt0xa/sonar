package cmd

import (
	"github.com/spf13/cobra"
)

func (c *Command) profileGet(cmd *cobra.Command) runFunc {
	cmd.Args = cobra.NoArgs

	return func(cmd *cobra.Command, args []string) error {
		out, err := c.svc.ProfileGet(cmd.Context())
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}
