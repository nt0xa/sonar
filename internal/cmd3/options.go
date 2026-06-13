package cmd3

import "github.com/spf13/cobra"

var defaultOptions = options{
	allowFileAccess: false,
	preExec:         nil,
}

type options struct {
	allowFileAccess bool
	preExec         func(root *cobra.Command)
}

// Option configures a Command.
type Option func(*options)

// AllowFileAccess enables flags that read local files (e.g. http --file). For use in
// the CLI application only — must stay false for messengers, otherwise there is a local
// file read vulnerability.
func AllowFileAccess(b bool) Option {
	return func(o *options) { o.allowFileAccess = b }
}

// PreExec runs against the cobra root after the tree is built. Use it to add
// subcommands/flags, set help/usage templates, or set PersistentPreRunE.
func PreExec(f func(root *cobra.Command)) Option {
	return func(o *options) { o.preExec = f }
}
