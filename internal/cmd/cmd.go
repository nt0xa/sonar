package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/google/shlex"
	"github.com/spf13/cobra"

	"github.com/russtone/sonar/internal/actions"
)

//go:generate go run ./internal/codegen/*.go -type cmd

func init() {
	cobra.EnableCommandSorting = false
	cobra.AddTemplateFuncs(sprig.TxtFuncMap())
}

type Command struct {
	actions actions.Actions
	options options
}

func New(a actions.Actions, opts ...Option) *Command {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	if a == nil && options.initActions == nil {
		panic("you will need to provide either actions != nil or options.initActions")
	}

	return &Command{
		actions: a,
		options: options,
	}

}

func (c *Command) root(onResult func(actions.Result) error) *cobra.Command {
	var root = &cobra.Command{
		Use:   "sonar",
		Short: "CLI to control sonar server",
	}

	// Main payloads commands
	root.AddCommand(c.withAuthCheck(c.PayloadsCreate(onResult)))
	root.AddCommand(c.withAuthCheck(c.PayloadsList(onResult)))
	root.AddCommand(c.withAuthCheck(c.PayloadsUpdate(onResult)))
	root.AddCommand(c.withAuthCheck(c.PayloadsDelete(onResult)))
	root.AddCommand(c.withAuthCheck(c.PayloadsClear(onResult)))

	// DNS
	dns := &cobra.Command{
		Use:   "dns",
		Short: "Manage DNS records",
	}

	dns.AddCommand(c.withAuthCheck(c.DNSRecordsCreate(onResult)))
	dns.AddCommand(c.withAuthCheck(c.DNSRecordsDelete(onResult)))
	dns.AddCommand(c.withAuthCheck(c.DNSRecordsList(onResult)))
	dns.AddCommand(c.withAuthCheck(c.DNSRecordsClear(onResult)))

	root.AddCommand(dns)

	// Events
	events := &cobra.Command{
		Use:   "events",
		Short: "Payloads events",
	}

	events.AddCommand(c.withAuthCheck(c.EventsList(onResult)))
	events.AddCommand(c.withAuthCheck(c.EventsGet(onResult)))

	root.AddCommand(events)

	// HTTP
	http := c.withAuthCheck(&cobra.Command{
		Use:   "http",
		Short: "Manage HTTP routes",
	})

	http.AddCommand(c.withAuthCheck(c.HTTPRoutesCreate(onResult)))
	http.AddCommand(c.withAuthCheck(c.HTTPRoutesDelete(onResult)))
	http.AddCommand(c.withAuthCheck(c.HTTPRoutesList(onResult)))
	http.AddCommand(c.withAuthCheck(c.HTTPRoutesClear(onResult)))

	root.AddCommand(http)

	// Profile
	root.AddCommand(c.withAuthCheck(c.ProfileGet(onResult)))

	// Users
	users := &cobra.Command{
		Use:   "users",
		Short: "Manage users",
	}

	users.AddCommand(c.withAdminCheck(c.UsersCreate(onResult)))
	users.AddCommand(c.withAdminCheck(c.UsersDelete(onResult)))

	root.AddCommand(users)

	return root
}

func (c *Command) Exec(ctx context.Context, args []string, onResult func(actions.Result) error) (string, string, error) {
	cmd := c.root(onResult)

	if c.options.preExec != nil {
		c.options.preExec(cmd)
	}

	cmd.SetArgs(args)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	// Disable print to error output.
	// Use result of cmd.ExecuteContext instead.
	cmd.SilenceErrors = true

	// Disable "Run 'sonar --help' for usage." messages.
	cmd.SilenceUsage = true

	// Late init actions.
	if c.actions == nil {
		persistentPreRunE := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if c.options.initActions == nil {
				return errors.New("actions are not initialized")
			}
			acts, err := c.options.initActions()
			if err != nil {
				return err
			}
			c.actions = acts

			if persistentPreRunE != nil {
				return persistentPreRunE(cmd, args)
			}

			return nil
		}
	}

	if err := cmd.ExecuteContext(ctx); err != nil {
		return "", "", err
	}

	return stdout.String(), stderr.String(), nil
}

func (c *Command) ParseAndExec(ctx context.Context, s string, onResult func(actions.Result) error) (string, string, error) {
	args, _ := shlex.Split(strings.TrimLeft(s, "/"))
	return c.Exec(ctx, args, onResult)
}
