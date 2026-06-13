package cmd2

import (
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/pkg/cmdx"
)

func (c *Command) profileGet(cmd *cobra.Command) cmdx.RunFunc {
	cmd.Args = cobra.NoArgs

	return func(cmd *cobra.Command, args []string) error {
		out, err := c.svc.ProfileGet(cmd.Context())
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}
