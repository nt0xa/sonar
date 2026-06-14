// Package cmd builds the sonar client CLI command tree on top of
// [service.Service] using spf13/cobra directly — no cmdx. Per-command flag /
// argument / validation wiring lives in this package next to the commands
// themselves via the closure pattern: a build func that configures the command
// and returns its run. Authorization (incl. admin-only commands) is enforced by
// the service layer, not here.
package cmd

import (
	"bytes"
	"context"
	"strings"

	"github.com/google/shlex"
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
)

func init() {
	// Keep commands in registration order in help output.
	cobra.EnableCommandSorting = false
}

// runFunc is cobra's RunE signature.
type runFunc = func(cmd *cobra.Command, args []string) error

// mainGroup is the display group top-level commands are tagged into.
const mainGroup = "main"

// Command builds the client command tree against a service.Service.
type Command struct {
	svc  service.Service
	opts options
}

// New returns a Command backed by svc.
func New(svc service.Service, opts ...Option) *Command {
	o := defaultOptions
	for _, opt := range opts {
		opt(&o)
	}
	return &Command{svc: svc, opts: o}
}

// command builds a runnable command from build, which wires the command's flags
// and returns its run closure.
func command(name, short string, build func(cmd *cobra.Command) runFunc) *cobra.Command {
	cmd := &cobra.Command{Use: name, Short: short}
	cmd.RunE = build(cmd)
	return cmd
}

// Root assembles the command tree. The whole tree lives here; each leaf
// references a build method (see the per-resource files) that wires its flags
// and returns its run closure.
func (c *Command) Root() *cobra.Command {
	root := &cobra.Command{Use: "sonar", Short: "CLI to control sonar server"}
	root.AddGroup(&cobra.Group{ID: mainGroup, Title: "Main commands"})

	// Payloads are the primary commands ("main" group); the rest are grouped into
	// containers (or left ungrouped) and show under cobra's "Additional Commands".
	for _, cmd := range []*cobra.Command{
		command("new", "Create a new payload", c.payloadsCreate),
		command("list", "List payloads", c.payloadsList),
		command("mod", "Modify existing payload", c.payloadsUpdate),
		command("del", "Delete payload", c.payloadsDelete),
		command("clr", "Delete multiple payloads", c.payloadsClear),
		command("profile", "Get current user info", c.profileGet),
	} {
		cmd.GroupID = mainGroup
		root.AddCommand(cmd)
	}

	dns := &cobra.Command{Use: "dns", Short: "Manage DNS records"}
	dns.AddCommand(
		command("new", "Create new DNS records", c.dnsRecordsCreate),
		command("del", "Delete DNS record", c.dnsRecordsDelete),
		command("list", "List DNS records", c.dnsRecordsList),
		command("clr", "Delete multiple DNS records", c.dnsRecordsClear),
	)
	root.AddCommand(dns)

	events := &cobra.Command{Use: "events", Short: "View events"}
	events.AddCommand(
		command("list", "List payload events", c.eventsList),
		command("get", "Get payload event by INDEX", c.eventsGet),
	)
	root.AddCommand(events)

	http := &cobra.Command{Use: "http", Short: "Manage HTTP routes"}
	http.AddCommand(
		command("new", "Create new HTTP route", c.httpRoutesCreate),
		command("mod", "Update HTTP route", c.httpRoutesUpdate),
		command("del", "Delete HTTP route", c.httpRoutesDelete),
		command("list", "List HTTP routes", c.httpRoutesList),
		command("clr", "Delete multiple HTTP routes", c.httpRoutesClear),
	)
	root.AddCommand(http)

	users := &cobra.Command{Use: "users", Short: "Manage users"}
	users.AddCommand(
		command("new", "Create new user", c.usersCreate),
		command("del", "Delete user", c.usersDelete),
	)
	root.AddCommand(users)

	audit := &cobra.Command{Use: "audit", Short: "View audit records"}
	audit.AddCommand(
		command("list", "List audit records", c.auditRecordsList),
		command("get", "Get audit record by ID", c.auditRecordsGet),
	)
	root.AddCommand(audit)

	return root
}

// Exec builds the tree, applies the PreExec hook, installs a result sink in ctx, and runs
// args. The result is either the typed service output from the invoked leaf or, when no leaf
// produced data (help/usage/completion), the raw text cobra wrote. Execution errors are
// returned separately.
func (c *Command) Exec(ctx context.Context, args []string) (any, error) {
	sink := &resultSink{}
	ctx = withSink(ctx, sink)

	root := c.Root()
	if c.opts.preExec != nil {
		c.opts.preExec(root)
	}

	// Capture cobra's stdout/stderr (help/usage/completion text) into one buffer.
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)

	// Errors are returned from Exec, not printed by cobra.
	root.SilenceErrors = true
	root.SilenceUsage = true

	if err := root.ExecuteContext(ctx); err != nil {
		return nil, err
	}

	if sink.out != nil {
		return sink.out, nil
	}

	return buf.String(), nil
}

// ParseAndExec splits s into args (stripping a leading "/", as messengers send) and runs it.
func (c *Command) ParseAndExec(ctx context.Context, s string) (any, error) {
	args, _ := shlex.Split(strings.TrimLeft(s, "/"))
	return c.Exec(ctx, args)
}
