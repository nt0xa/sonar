package cmdx

import (
	"slices"

	"github.com/spf13/cobra"
)

// RunFunc is cobra's RunE signature.
type RunFunc func(cmd *cobra.Command, args []string) error

// Wrapper wraps a RunFunc — pre/post/short-circuit around execution.
type Wrapper func(RunFunc) RunFunc

// mainGroupID is the default display group New tags top-level commands into.
const mainGroupID = "main"

// Command is a node in the tree. It attaches children to cmd and applies the
// accumulated wrappers to every leaf registered through it. groupID, when set,
// tags registered commands into a cobra display group.
type Command struct {
	cmd      *cobra.Command
	wrappers []Wrapper
	groupID  string
}

// New creates a root command with a default display group so every top-level
// command added through it is grouped under one heading.
func New(name, short string, w ...Wrapper) *Command {
	root := &cobra.Command{Use: name, Short: short}
	root.AddGroup(&cobra.Group{ID: mainGroupID, Title: "Commands"})
	return &Command{cmd: root, wrappers: w, groupID: mainGroupID}
}

// Cobra returns the underlying root command (call Execute / preExec / set
// PersistentPreRunE / templates on it).
func (c *Command) Cobra() *cobra.Command { return c.cmd }

// Wrap appends wrappers to this scope.
func (c *Command) Wrap(w ...Wrapper) { c.wrappers = append(c.wrappers, w...) }

// AddCommand registers a prebuilt cobra command (e.g. a generated action command),
// wrapping its existing RunE/Run with the accumulated wrappers and tagging it into
// the current display group.
func (c *Command) AddCommand(sub *cobra.Command) {
	switch {
	case sub.RunE != nil:
		sub.RunE = c.apply(sub.RunE)
	case sub.Run != nil:
		run := sub.Run
		sub.Run = nil
		sub.RunE = c.apply(func(cmd *cobra.Command, args []string) error {
			run(cmd, args)
			return nil
		})
	}
	if c.groupID != "" {
		sub.GroupID = c.groupID
	}
	c.cmd.AddCommand(sub)
}

// Add is sugar: build a leaf from name/short/run (+opts for flags/args) and
// register it through AddCommand.
func (c *Command) Add(name, short string, run RunFunc, opts ...func(*cobra.Command)) {
	sub := &cobra.Command{Use: name, Short: short, RunE: run}
	for _, opt := range opts {
		opt(sub)
	}
	c.AddCommand(sub)
}

// Group adds a non-runnable container subcommand and registers its children via
// fn, which inherits this scope's (cloned) wrappers.
func (c *Command) Group(name, short string, fn func(c *Command), opts ...func(*cobra.Command)) {
	sub := &cobra.Command{Use: name, Short: short}
	for _, opt := range opts {
		opt(sub)
	}
	if c.groupID != "" {
		sub.GroupID = c.groupID
	}
	c.cmd.AddCommand(sub)
	fn(&Command{cmd: sub, wrappers: slices.Clone(c.wrappers)})
}

func (c *Command) apply(run RunFunc) RunFunc {
	out := run
	ws := slices.Clone(c.wrappers)
	slices.Reverse(ws) // first registered = outermost, like webx
	for _, w := range ws {
		out = w(out)
	}
	return out
}
