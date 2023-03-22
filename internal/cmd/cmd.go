package cmd

import (
	"bytes"
	"context"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/google/shlex"
	"github.com/spf13/cobra"

	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/utils/errors"
)

//go:generate go run ./internal/codegen/*.go -type cmd

func init() {
	cobra.EnableCommandSorting = false
	cobra.AddTemplateFuncs(sprig.TxtFuncMap())
}

type Command struct {
	actions actions.Actions
	handler actions.ResultHandler

	preExec func(context.Context, *cobra.Command)
	local   bool
}

type CommandOption func(*Command)

func PreExec(f func(context.Context, *cobra.Command)) func(*Command) {
	return func(c *Command) {
		c.preExec = f
	}
}

func Local() func(*Command) {
	return func(c *Command) {
		c.local = true
	}
}

func New(acts actions.Actions, h actions.ResultHandler, opts ...CommandOption) *Command {
	c := &Command{
		actions: acts,
		handler: h,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Command) root(user *actions.User) *cobra.Command {
	var root = &cobra.Command{
		Use:   "sonar",
		Short: "CLI to control sonar server",
	}

	// There is no access control inside commands,
	// so if user is not allowed to do command we just
	// don't add it to root.

	// Currently, there are no default commands available
	// for unauthorized users, but some controller can implement
	// their own unauthorized commands.
	if user == nil {
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
	root.AddCommand(c.ProfileGet())

	// Users
	if user.IsAdmin {
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

func (c *Command) Exec(ctx context.Context, args []string) {

	profile, err := c.actions.ProfileGet(ctx)
	if err != nil {
		c.handler.OnResult(ctx, actions.ErrorResult{err})
		return
	}

	root := c.root(&profile.User)

	if c.preExec != nil {
		c.preExec(ctx, root)
	}

	root.SetArgs(args)

	bb := &bytes.Buffer{}
	root.SetErr(bb)
	root.SetOut(bb)

	// There is no subcommands which means that user is unauthorized
	// and no commands available for unauthorized users in current controller.
	if !root.HasAvailableSubCommands() {
		c.handler.OnResult(ctx, actions.ErrorResult{errors.Unauthorized()})
		return
	}

	if err := root.ExecuteContext(ctx); err != nil {
		c.handler.OnResult(ctx, actions.ErrorResult{errors.Internal(err)})
		return
	}

	if bb.String() != "" {
		c.handler.OnResult(ctx, actions.TextResult{bb.String()})
	}
}

func (c *Command) ParseAndExec(ctx context.Context, s string) {
	args, _ := shlex.Split(strings.TrimLeft(s, "/"))
	c.Exec(ctx, args)
}

func DefaultMessengersPreExec(ctx context.Context, root *cobra.Command) {
	root.SetHelpTemplate(helpTemplate)
	root.SetUsageTemplate(usageTemplate)
	root.CompletionOptions = cobra.CompletionOptions{
		DisableDefaultCmd: true,
	}
}

var helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}
{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

var usageTemplate = `
Usage:{{if .Runnable}}{{if .HasParent}}
  {{.UseLine | replace "sonar " "/"}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
	{{if .HasParent}}{{.CommandPath | replace "sonar " "/"}} {{else}}/{{end}}[command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  /{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{if .HasParent}}{{.CommandPath | replace "sonar " "/"}} {{else}}/{{end}}[command] --help" for more information about a command.{{end}}`
