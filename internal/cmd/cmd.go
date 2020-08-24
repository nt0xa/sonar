package cmd

import (
	"bytes"
	"context"

	"github.com/Masterminds/sprig"
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func init() {
	cobra.EnableCommandSorting = false
	cobra.AddTemplateFuncs(sprig.TxtFuncMap())
}

type ResultHandler func(context.Context, interface{})
type PreExecFunc func(*cobra.Command, *models.User)

type Command struct {
	Actions       actions.Actions
	ResultHandler ResultHandler
	PreExec       PreExecFunc
}

func (c *Command) Root(u *models.User) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sonarctl",
		Short: "CLI to control your sonar server",
	}

	// There is no access control inside commands,
	// so if user is not allowed to do command we just
	// don't add it to root.

	// Currenty, threre are no default commands available
	// for unauthorized users, but some controller can implement
	// their own unauthorized commands and add this commands to root
	// using `preExec`.
	if u == nil {
		return cmd
	}

	// Main payloads commands
	cmd.AddCommand(c.CreatePayload())
	cmd.AddCommand(c.UpdatePayload())
	cmd.AddCommand(c.DeletePayload())
	cmd.AddCommand(c.ListPayloads())

	cmd.AddCommand(c.DNS())

	if u.IsAdmin {
		cmd.AddCommand(c.Users())
	}

	return cmd
}

func (c *Command) Exec(ctx context.Context, user *models.User, args []string) (string, errors.Error) {
	root := c.Root(user)

	if c.PreExec != nil {
		c.PreExec(root, user)
	}

	root.SetArgs(args)

	bb := &bytes.Buffer{}
	root.SetErr(bb)
	root.SetOut(bb)

	ctx = actions.SetUser(ctx, user)

	// There is no subcommands which means that user is unauthorized
	// and no commands available for unauthorized users in current controller.
	if !root.HasAvailableSubCommands() {
		return "", errors.Unauthorized()
	}

	if err := root.ExecuteContext(ctx); err != nil {
		e, ok := err.(errors.Error)
		if !ok {
			return bb.String(), errors.Internal(err)
		}

		return bb.String(), e
	}

	return bb.String(), nil
}
