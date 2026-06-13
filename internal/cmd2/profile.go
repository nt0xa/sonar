package cmd2

import (
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/pkg/cmdx"
)

func (c *Command) addProfile(g *cmdx.Command) {
	p := &profileGet{c: c}
	g.Add("profile", "Get current user info", p.run, p.flags)
}

type profileGet struct {
	c *Command
}

func (x *profileGet) flags(cmd *cobra.Command) {
	cmd.Args = cobra.NoArgs
}

func (x *profileGet) run(cmd *cobra.Command, args []string) error {
	out, err := x.c.svc.ProfileGet(cmd.Context())
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}
