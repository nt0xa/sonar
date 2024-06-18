package cmd

import (
	"github.com/nt0xa/sonar/internal/actions"
	"github.com/spf13/cobra"
)

var defaultOptions = options{
	allowFileAccess: false,
	preExec:         nil,
	initActions:     nil,
}

type options struct {
	allowFileAccess bool
	preExec         func(*cobra.Command)
	initActions     func() (actions.Actions, error)
}

type Option func(*options)

// AllowFileAccess enables arguments in some commands that use local files.
// Must be for use in CLI application only, not messengers.
// Otherwise there will be local file read vulnerability.
func AllowFileAccess(b bool) Option {
	return func(opts *options) {
		opts.allowFileAccess = b
	}
}

// PreExec is called after Root command is created.
// Should be used to add specific subcommands or flags (e.g. "--json" for CLI).
func PreExec(f func(*cobra.Command)) Option {
	return func(opts *options) {
		opts.preExec = f
	}
}

// InitActions is a workaround for CLI.
// In the CLI we can't pass `action.Actions`, before we parse config and flags,
// so we need a way to create actions late.
func InitActions(f func() (actions.Actions, error)) Option {
	return func(opts *options) {
		opts.initActions = f
	}
}
