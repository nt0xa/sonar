package cmd

import (
	"github.com/nt0xa/sonar/internal/actions"
	"github.com/spf13/cobra"
)

var defaultOptions = options{
	allowFileAccess: false,
	preExec:         nil,
}

type options struct {
	allowFileAccess bool
	preExec         func(actions *actions.Actions, cmd *cobra.Command)
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
func PreExec(f func(actions *actions.Actions, cmd *cobra.Command)) Option {
	return func(opts *options) {
		opts.preExec = f
	}
}
