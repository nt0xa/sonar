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
	Root(*actions.User) *cobra.Command
	Exec(context.Context, *actions.User, []string) (string, errors.Error)
}

func New(actions actions.Actions, handler actions.ResultHandler, preExec PreExec) Command {
	return &command{
		actions: actions,
		handler: handler,
		preExec: preExec,
	}
}

func (c *command) Root(u *actions.User) *cobra.Command {
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
	root.AddCommand(c.PayloadsCreate())
	root.AddCommand(c.PayloadsList())
	root.AddCommand(c.PayloadsUpdate())
	root.AddCommand(c.PayloadsDelete())

	// DNS
	dns := &cobra.Command{
		Use:   "dns",
		Short: "Manage DNS records",
	}

	dns.AddCommand(c.DNSRecordsCreate())
	dns.AddCommand(c.DNSRecordsDelete())
	dns.AddCommand(c.DNSRecordsList())

	root.AddCommand(dns)

	// Events
	events := &cobra.Command{
		Use:   "events",
		Short: "Payloads events",
	}

	events.AddCommand(c.EventsList())
	events.AddCommand(c.EventsGet())

	root.AddCommand(events)

	// HTTP
	http := &cobra.Command{
		Use:   "http",
		Short: "Manage HTTP routes",
	}

	http.AddCommand(c.HTTPRoutesCreate())
	http.AddCommand(c.HTTPRoutesDelete())
	http.AddCommand(c.HTTPRoutesList())

	root.AddCommand(http)

	// User
	root.AddCommand(c.UserCurrent())

	// Users
	if u.IsAdmin {
		users := &cobra.Command{
			Use:   "users",
			Short: "Manage users",
		}

		users.AddCommand(c.UsersCreate())
		users.AddCommand(c.UsersDelete())

		root.AddCommand(users)
	}

	return root
}

func (c *command) Exec(ctx context.Context, u *actions.User, args []string) (string, errors.Error) {
	root := c.Root(u)

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
