// Package cmd2 builds the sonar client CLI command tree on top of
// [service.Service] (instead of the legacy actions.Actions) using the [cmdx]
// command-tree builder. Per-command flag/argument/validation wiring lives in this
// package next to the commands themselves — there is no codegen.
package cmd2

import (
	"bytes"
	"context"
	"strings"

	"github.com/google/shlex"
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/pkg/cmdx"
)

func init() {
	// Keep commands in registration order in help output.
	cobra.EnableCommandSorting = false
}

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

// Root assembles the command tree and returns the cmdx node. Callers may grab
// .Cobra() to attach persistent flags / pre-run hooks before executing. The whole
// tree lives here; each leaf references a build method (see the per-resource files)
// that wires its flags and returns its run closure.
func (c *Command) Root() *cmdx.Command {
	// authWrapper is the base wrapper: every command requires authentication and
	// gets the resolved profile stashed in context (see auth.go).
	root := cmdx.New("sonar", "CLI to control sonar server", c.authWrapper)

	// Payloads + profile live at the top level ("main" group).
	root.Cmd("new", "Create a new payload", c.payloadsCreate)
	root.Cmd("list", "List payloads", c.payloadsList)
	root.Cmd("mod", "Modify existing payload", c.payloadsUpdate)
	root.Cmd("del", "Delete payload", c.payloadsDelete)
	root.Cmd("clr", "Delete multiple payloads", c.payloadsClear)
	root.Cmd("profile", "Get current user info", c.profileGet)

	root.Group("dns", "Manage DNS records", func(g *cmdx.Command) {
		g.Cmd("new", "Create new DNS records", c.dnsRecordsCreate)
		g.Cmd("del", "Delete DNS record", c.dnsRecordsDelete)
		g.Cmd("list", "List DNS records", c.dnsRecordsList)
		g.Cmd("clr", "Delete multiple DNS records", c.dnsRecordsClear)
	})

	root.Group("events", "View events", func(g *cmdx.Command) {
		g.Cmd("list", "List payload events", c.eventsList)
		g.Cmd("get", "Get payload event by INDEX", c.eventsGet)
	})

	root.Group("http", "Manage HTTP routes", func(g *cmdx.Command) {
		g.Cmd("new", "Create new HTTP route", c.httpRoutesCreate)
		g.Cmd("mod", "Update HTTP route", c.httpRoutesUpdate)
		g.Cmd("del", "Delete HTTP route", c.httpRoutesDelete)
		g.Cmd("list", "List HTTP routes", c.httpRoutesList)
		g.Cmd("clr", "Delete multiple HTTP routes", c.httpRoutesClear)
	})

	// Admin-only groups stack adminWrapper on top of authWrapper.
	root.Group("users", "Manage users", func(g *cmdx.Command) {
		g.Wrap(c.adminWrapper)
		g.Cmd("new", "Create new user", c.usersCreate)
		g.Cmd("del", "Delete user", c.usersDelete)
	})

	root.Group("audit", "View audit records", func(g *cmdx.Command) {
		g.Wrap(c.adminWrapper)
		g.Cmd("list", "List audit records", c.auditRecordsList)
		g.Cmd("get", "Get audit record by ID", c.auditRecordsGet)
	})

	return root
}

// Exec builds the tree, applies the PreExec hook, installs a result sink in ctx, and runs
// args. The result is either the typed service output from the invoked leaf or, when no leaf
// produced data (help/usage/completion), the raw text cobra wrote. Execution errors are
// returned separately.
func (c *Command) Exec(ctx context.Context, args []string) (any, error) {
	sink := &resultSink{}
	ctx = withSink(ctx, sink)

	root := c.Root().Cobra()
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
