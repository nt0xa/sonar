package cmd

import (
	"bytes"
	"context"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/google/shlex"
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/actions"
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

	return &Command{
		actions: a,
		options: options,
	}
}

func (c *Command) root(onResult func(context.Context, actions.Result) error) *cobra.Command {
	var root = &cobra.Command{
		Use:   "sonar",
		Short: "CLI to control sonar server",
	}

	root.AddGroup(
		&cobra.Group{
			ID:    "main",
			Title: "Main commands",
		},
	)

	// Main payloads commands
	for _, cmd := range []*cobra.Command{
		c.PayloadsCreate(onResult),
		c.PayloadsList(onResult),
		c.PayloadsUpdate(onResult),
		c.PayloadsDelete(onResult),
		c.PayloadsClear(onResult),
	} {
		cmd.GroupID = "main"
		root.AddCommand(c.withAuthCheck(cmd))
	}

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
		Short: "View events",
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
	http.AddCommand(c.withAuthCheck(c.HTTPRoutesUpdate(onResult)))
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

func (c *Command) Exec(
	ctx context.Context,
	args []string,
	onResult func(context.Context, actions.Result) error,
) (string, string, error) {
	cmd := c.root(onResult)

	if c.options.preExec != nil {
		c.options.preExec(&c.actions, cmd)
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

	if err := cmd.ExecuteContext(ctx); err != nil {
		return "", "", err
	}

	return stdout.String(), stderr.String(), nil
}

func (c *Command) ParseAndExec(
	ctx context.Context,
	s string,
	onResult func(context.Context, actions.Result) error,
) (string, string, error) {
	args, _ := shlex.Split(strings.TrimLeft(s, "/"))
	return c.Exec(ctx, args, onResult)
}
