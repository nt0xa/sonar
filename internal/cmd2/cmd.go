// Package cmd2 builds the sonar client CLI command tree on top of
// [service.Service] (instead of the legacy actions.Actions) using the [cmdx]
// command-tree builder. Per-command flag/argument/validation wiring lives in this
// package next to the commands themselves — there is no codegen.
package cmd2

import (
	"context"

	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/pkg/cmdx"
)

// Command builds the client command tree against a service.Service.
type Command struct {
	svc service.Service
}

// New returns a Command backed by svc.
func New(svc service.Service) *Command {
	return &Command{svc: svc}
}

// Root assembles the command tree and returns the cmdx node. Callers may grab
// .Cobra() to attach persistent flags / pre-run hooks before executing.
func (c *Command) Root() *cmdx.Command {
	// authWrapper is the base wrapper: every command requires authentication and
	// gets the resolved profile stashed in context (see auth.go).
	root := cmdx.New("sonar", "CLI to control sonar server", c.authWrapper)

	// Payloads live at the top level ("main" group).
	c.addPayloads(root)

	root.Group("dns", "Manage DNS records", c.addDNS)
	root.Group("events", "View events", c.addEvents)
	root.Group("http", "Manage HTTP routes", c.addHTTP)

	c.addProfile(root)

	// Admin-only groups stack adminWrapper on top of authWrapper.
	root.Group("users", "Manage users", func(g *cmdx.Command) {
		g.Wrap(c.adminWrapper)
		c.addUsers(g)
	})
	root.Group("audit", "View audit records", func(g *cmdx.Command) {
		g.Wrap(c.adminWrapper)
		c.addAudit(g)
	})

	return root
}

// Exec builds the tree, installs a result sink in ctx, runs args, and returns the
// single result produced by the invoked leaf (nil for help/completion).
func (c *Command) Exec(ctx context.Context, args []string) (any, error) {
	sink := &resultSink{}
	ctx = withSink(ctx, sink)

	root := c.Root().Cobra()
	root.SetArgs(args)

	// Errors are returned from Exec, not printed by cobra.
	root.SilenceErrors = true
	root.SilenceUsage = true

	if err := root.ExecuteContext(ctx); err != nil {
		return nil, err
	}

	return sink.out, nil
}
