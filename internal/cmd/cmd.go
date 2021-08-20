package cmd

import (
	"bytes"
	"context"

	"github.com/Masterminds/sprig"
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

//go:generate go run ./internal/codegen/*.go -type cmd

func init() {
	cobra.EnableCommandSorting = false
	cobra.AddTemplateFuncs(sprig.TxtFuncMap())
}

type ResultHandler func(context.Context, interface{})
type PreExec func(*cobra.Command, *actions.User)

type command struct {
	actions actions.Actions
	handler actions.ResultHandler
	preExec PreExec
}

type Command interface {
	Root(*actions.User, bool) *cobra.Command
	Exec(context.Context, *actions.User, bool, []string) (string, errors.Error)
}

func New(actions actions.Actions, handler actions.ResultHandler, preExec PreExec) Command {
	return &command{
		actions: actions,
		handler: handler,
		preExec: preExec,
	}
}

func (c *command) Root(u *actions.User, local bool) *cobra.Command {
	var root = &cobra.Command{
		Use:   "sonar",
		Short: "CLI to control sonar server",
	}

	// There is no access control inside commands,
	// so if user is not allowed to do command we just
	// don't add it to root.

	// Currently, there are no default commands available
	// for unauthorized users, but some controller can implement
	// their own unauthorized commands and add this commands to root
	// using `preExec`.
	if u == nil {
		return root
	}

	// Main payloads commands
	root.AddCommand(c.PayloadsCreate(local))
	root.AddCommand(c.PayloadsList(local))
	root.AddCommand(c.PayloadsUpdate(local))
	root.AddCommand(c.PayloadsDelete(local))

	// DNS
	dns := &cobra.Command{
		Use:   "dns",
		Short: "Manage DNS records",
	}

	dns.AddCommand(c.DNSRecordsCreate(local))
	dns.AddCommand(c.DNSRecordsDelete(local))
	dns.AddCommand(c.DNSRecordsList(local))

	root.AddCommand(dns)

	// Events
	events := &cobra.Command{
		Use:   "events",
		Short: "Payloads events",
	}

	events.AddCommand(c.EventsList(local))
	events.AddCommand(c.EventsGet(local))

	root.AddCommand(events)

	// HTTP
	http := &cobra.Command{
		Use:   "http",
		Short: "Manage HTTP routes",
	}

	http.AddCommand(c.HTTPRoutesCreate(local))
	http.AddCommand(c.HTTPRoutesDelete(local))
	http.AddCommand(c.HTTPRoutesList(local))

	root.AddCommand(http)

	// User
	root.AddCommand(c.UserCurrent(local))

	// Users
	if u.IsAdmin {
		users := &cobra.Command{
			Use:   "users",
			Short: "Manage users",
		}

		users.AddCommand(c.UsersCreate(local))
		users.AddCommand(c.UsersDelete(local))

		root.AddCommand(users)
	}

	return root
}

func (c *command) Exec(ctx context.Context, u *actions.User, local bool, args []string) (string, errors.Error) {
	root := c.Root(u, local)

	if c.preExec != nil {
		c.preExec(root, u)
	}

	root.SetArgs(args)

	bb := &bytes.Buffer{}
	root.SetErr(bb)
	root.SetOut(bb)

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
