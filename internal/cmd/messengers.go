package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	// The derived messenger usage template pipes dynamic command paths through
	// `replace` to turn "sonar dns" into "/dns"; cobra has no such built-in func.
	cobra.AddTemplateFunc("replace", func(old, new, src string) string {
		return strings.NewReplacer(old, new).Replace(src)
	})
}

// The messenger help/usage templates are derived from cobra's defaults by string
// replacement instead of being carried as full hand-written templates. This couples to
// cobra's default template text: if a cobra upgrade changes the substrings below, a
// replacement silently no-ops — TestDefaultMessengersPreExec guards against that.
var (
	messengersUsageTemplate = strings.NewReplacer(
		// Slash-style "Usage:" header: drop the root's own use line, render "/path".
		"{{if .Runnable}}\n  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}\n  {{.CommandPath}} [command]{{end}}",
		"{{if .Runnable}}{{if .HasParent}}\n  {{.UseLine | replace \"sonar \" \"/\"}}{{end}}{{end}}{{if .HasAvailableSubCommands}}\n\t{{if .HasParent}}{{.CommandPath | replace \"sonar \" \"/\"}} {{else}}/{{end}}[command]{{end}}",
		// Prefix every listed command name with "/".
		"\n  {{rpad .Name .NamePadding }} {{.Short}}",
		"\n  /{{rpad .Name .NamePadding }} {{.Short}}",
		// Slash-style trailing "Use ... --help" line.
		"Use \"{{.CommandPath}} [command] --help\" for more information about a command.",
		"Use \"{{if .HasParent}}{{.CommandPath | replace \"sonar \" \"/\"}} {{else}}/{{end}}[command] --help\" for more information about a command.",
	).Replace((&cobra.Command{}).UsageTemplate())

	messengersHelpTemplate = strings.NewReplacer(
		// Drop the blank line between the long/short description and the usage block.
		"{{. | trimTrailingWhitespaces}}\n\n{{end}}",
		"{{. | trimTrailingWhitespaces}}\n{{end}}",
	).Replace((&cobra.Command{}).HelpTemplate())
)

// DefaultMessengersPreExec adapts the command tree for messenger (chat) use: it renders
// commands in slash style (e.g. "/dns" instead of "sonar dns") and drops the auto-generated
// completion command, which is meaningless in chat. Pass it to New via PreExec.
func DefaultMessengersPreExec(root *cobra.Command) {
	root.SetHelpTemplate(messengersHelpTemplate)
	root.SetUsageTemplate(messengersUsageTemplate)
	root.CompletionOptions.DisableDefaultCmd = true
}
