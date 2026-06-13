package cmd2

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/pkg/cmdx"
)

type profileKeyType struct{}

var profileKey = profileKeyType{}

// authWrapper requires a valid caller and stashes the resolved profile in context
// so a stacked adminWrapper can reuse it without a second ProfileGet.
func (c *Command) authWrapper(next cmdx.RunFunc) cmdx.RunFunc {
	return func(cmd *cobra.Command, args []string) error {
		p, err := c.svc.ProfileGet(cmd.Context())
		if err != nil {
			return err
		}
		cmd.SetContext(context.WithValue(cmd.Context(), profileKey, p))
		return next(cmd, args)
	}
}

// adminWrapper rejects non-admin callers. It relies on authWrapper (registered
// first, so outermost) having stashed the profile.
func (c *Command) adminWrapper(next cmdx.RunFunc) cmdx.RunFunc {
	return func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Context().Value(profileKey).(*service.User)
		if p == nil || !p.IsAdmin {
			return service.Forbiddenf("admin only")
		}
		return next(cmd, args)
	}
}
